package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/saas-zero/saas-zero-common/pkg/envconf"
	"strconv"

	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/config"
	sysapisServer "github.com/saas-zero/saas-zero-basedata/rpc/internal/server/sysapis"
	sysdeptsServer "github.com/saas-zero/saas-zero-basedata/rpc/internal/server/sysdepts"
	sysdictdatasServer "github.com/saas-zero/saas-zero-basedata/rpc/internal/server/sysdictdatas"
	sysdictsServer "github.com/saas-zero/saas-zero-basedata/rpc/internal/server/sysdicts"
	sysinitServer "github.com/saas-zero/saas-zero-basedata/rpc/internal/server/sysinit"
	syslogsServer "github.com/saas-zero/saas-zero-basedata/rpc/internal/server/syslogs"
	sysmenusServer "github.com/saas-zero/saas-zero-basedata/rpc/internal/server/sysmenus"
	syspackagesServer "github.com/saas-zero/saas-zero-basedata/rpc/internal/server/syspackages"
	sysrolesServer "github.com/saas-zero/saas-zero-basedata/rpc/internal/server/sysroles"
	systenantsServer "github.com/saas-zero/saas-zero-basedata/rpc/internal/server/systenants"
	sysusersServer "github.com/saas-zero/saas-zero-basedata/rpc/internal/server/sysusers"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if uid := md.Get("x-user-id"); len(uid) > 0 {
			if id, err := strconv.ParseInt(uid[0], 10, 64); err == nil {
				ctx = mixins.SetCurrentUserId(ctx, id)
			}
		}
		if uname := md.Get("x-user-name"); len(uname) > 0 {
			ctx = mixins.SetCurrentUserName(ctx, uname[0])
		}
		if tid := md.Get("x-tenant-id"); len(tid) > 0 {
			if id, err := strconv.ParseInt(tid[0], 10, 64); err == nil {
				ctx = mixins.SetCurrentTenantId(ctx, id)
			}
		}
	}
	return handler(ctx, req)
}

var configFile = flag.String("f", "etc/basedataService.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 环境变量覆盖：POSTGRES_DSN 优先生效（生产用），无则用 YAML 明文（本地调试）
	c.Postgres.DataSource = envconf.String("POSTGRES_DSN", c.Postgres.DataSource)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		apps.RegisterSysUsersServer(grpcServer, sysusersServer.NewSysUsersServer(ctx))
		apps.RegisterSysRolesServer(grpcServer, sysrolesServer.NewSysRolesServer(ctx))
		apps.RegisterSysDeptsServer(grpcServer, sysdeptsServer.NewSysDeptsServer(ctx))
		apps.RegisterSysMenusServer(grpcServer, sysmenusServer.NewSysMenusServer(ctx))
		apps.RegisterSysDictsServer(grpcServer, sysdictsServer.NewSysDictsServer(ctx))
		apps.RegisterSysDictDatasServer(grpcServer, sysdictdatasServer.NewSysDictDatasServer(ctx))
		apps.RegisterSysTenantsServer(grpcServer, systenantsServer.NewSysTenantsServer(ctx))
		apps.RegisterSysPackagesServer(grpcServer, syspackagesServer.NewSysPackagesServer(ctx))
		apps.RegisterSysApisServer(grpcServer, sysapisServer.NewSysApisServer(ctx))
		apps.RegisterSysLogsServer(grpcServer, syslogsServer.NewSysLogsServer(ctx))
		apps.RegisterSysInitServer(grpcServer, sysinitServer.NewSysInitServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(authInterceptor)
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
