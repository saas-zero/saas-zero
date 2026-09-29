package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/config"
	"github.com/saas-zero/saas-zero-job/rpc/internal/server"
	"github.com/saas-zero/saas-zero-job/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
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

var configFile = flag.String("f", "etc/jobservice.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		apps.RegisterSysJobsServer(grpcServer, server.NewSysJobsServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(authInterceptor)
	defer s.Stop()

	fmt.Printf("Starting job rpc server at %s...\n", c.ListenOn)
	s.Start()
}
