package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strconv"

	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/envconf"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/config"
	"github.com/saas-zero/saas-zero-job/rpc/internal/server"
	"github.com/saas-zero/saas-zero-job/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// authInterceptor 从 gRPC metadata 注入审计用户上下文（x-user-id/x-user-name）
// job 为平台级服务：不注入 tenant（无 TenantMixin），
// 但 created_id/updated_id 等审计字段需要用户，缺失时由 logic 层兜底 system 用户。
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = withAuditUser(ctx)
	return handler(ctx, req)
}

// withAuditUser 从 metadata 读取用户，缺失则注入 system 用户（调度器/初始化场景）
func withAuditUser(ctx context.Context) context.Context {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if uid := md.Get("x-user-id"); len(uid) > 0 {
			if id, err := strconv.ParseInt(uid[0], 10, 64); err == nil {
				ctx = mixins.SetCurrentUserId(ctx, id)
			}
		}
		if uname := md.Get("x-user-name"); len(uname) > 0 {
			ctx = mixins.SetCurrentUserName(ctx, uname[0])
		}
	}
	if mixins.GetCurrentUserId(ctx) == 0 {
		ctx = mixins.SetCurrentUserId(ctx, 1)
	}
	if mixins.GetCurrentUserName(ctx) == "" {
		ctx = mixins.SetCurrentUserName(ctx, "system")
	}
	return ctx
}

// errnoInterceptor 把业务错误（*errno.Errno）转成带 gRPC 状态码的错误。
// gRPC 原生不携带自定义业务码，默认全部落到 codes.Unknown，API 层只能看到
// “内部服务器错误”；按业务码映射到状态码后，API 层才能还原出 {code,msg}。
func errnoInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}
	var e *errno.Errno
	if errors.As(err, &e) {
		return nil, status.Error(errnoToGRPC(e.Code), e.Msg)
	}
	return nil, err
}

// errnoToGRPC 业务码 → gRPC 状态码（只映射客户端可直接展示的错误码；
// 500 等内部错误映射为 codes.Internal，由 API 层统一隐藏细节）。
func errnoToGRPC(code int) codes.Code {
	switch {
	case code == errno.Unauthorized.Code:
		return codes.Unauthenticated
	case code == errno.Forbidden.Code:
		return codes.PermissionDenied
	case code >= 400 && code < 500:
		return codes.InvalidArgument
	case code >= 1000 && code < 1100:
		return codes.InvalidArgument
	default:
		return codes.Internal
	}
}

var configFile = flag.String("f", "etc/jobservice.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 生产环境用环境变量覆盖 YAML 明文敏感项（YAML 仅供本地调试）
	c.Postgres.DataSource = envconf.String("POSTGRES_DSN", c.Postgres.DataSource)
	c.CacheRedis.Host = envconf.String("REDIS_HOST", c.CacheRedis.Host)
	c.CacheRedis.Pass = envconf.String("REDIS_PASS", c.CacheRedis.Pass)
	if v := envconf.String("REDIS_DB", ""); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			c.CacheRedis.DB = db
		}
	}

	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		apps.RegisterSysJobsServer(grpcServer, server.NewSysJobsServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(authInterceptor, errnoInterceptor)
	defer s.Stop()

	fmt.Printf("Starting job rpc server at %s...\n", c.ListenOn)
	s.Start()
}
