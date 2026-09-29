// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"os"

	"github.com/saas-zero/saas-zero-auth/api/internal/config"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-common/pkg/redis"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	Redis      *redis.Client
	SysUsers   apps.SysUsersClient
	SysTenants apps.SysTenantsClient
	SysMenus   apps.SysMenusClient
	SysApis    apps.SysApisClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	cli := zrpc.MustNewClient(c.BaseDataRpc)
	rds, err := redis.NewClient(c.Redis)
	if err != nil {
		logx.Errorf("failed to init redis: %v", err)
		os.Exit(1)
	}
	return &ServiceContext{
		Config:     c,
		Redis:      rds,
		SysUsers:   apps.NewSysUsersClient(cli.Conn()),
		SysTenants: apps.NewSysTenantsClient(cli.Conn()),
		SysMenus:   apps.NewSysMenusClient(cli.Conn()),
		SysApis:    apps.NewSysApisClient(cli.Conn()),
	}
}
