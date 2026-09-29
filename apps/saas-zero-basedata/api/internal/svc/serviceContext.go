package svc

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"

	casbinapi "github.com/casbin/casbin/v2"
	"github.com/saas-zero/saas-zero-basedata/api/internal/config"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	commcasbin "github.com/saas-zero/saas-zero-common/pkg/casbin"
	"github.com/saas-zero/saas-zero-common/pkg/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	Redis        *redis.Client
	SysUsers     apps.SysUsersClient
	SysRoles     apps.SysRolesClient
	SysDepts     apps.SysDeptsClient
	SysMenus     apps.SysMenusClient
	SysDicts     apps.SysDictsClient
	SysDictDatas apps.SysDictDatasClient
	SysTenants   apps.SysTenantsClient
	SysPackages  apps.SysPackagesClient
	SysApis      apps.SysApisClient
	SysLogs      apps.SysLogsClient
	SysInit      apps.SysInitClient
	Enforcer     *casbinapi.SyncedEnforcer
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := zrpc.MustNewClient(c.Basedata, zrpc.WithUnaryClientInterceptor(authClientInterceptor))

	// Casbin enforcer 初始化：fail-closed。
	// 生产环境数据库/Enforcer 初始化失败即中止启动，绝不放行 /system/* 权限校验。
	var enf *casbinapi.SyncedEnforcer
	if !c.CasbinDisabled {
		db, err := sql.Open("postgres", c.CasbinPostgres.DataSource)
		if err != nil {
			log.Fatalf("fatal: failed to open casbin db: %v (fail-closed)", err)
		}
		enf, err = commcasbin.NewEnforcer(db, "casbin_rule")
		if err != nil {
			log.Fatalf("fatal: failed to init casbin enforcer: %v (fail-closed)", err)
		}
		// Background goroutine: periodically reload Casbin policies from DB.
		// Policies are updated by basedata-rpc's AssignApis RPC.
		// 30s interval for faster policy propagation during development.
		// 重载失败仅记录告警，继续使用最后一次成功加载的策略。
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if err := enf.LoadPolicy(); err != nil {
					log.Printf("CASBIN ALERT: reload policy error (keeping last-known policies): %v", err)
				}
			}
		}()
	}

	// Redis 客户端初始化：fail-closed。
	// 生产环境 Redis 不可用即中止启动，JWT 校验不得退化为仅验证签名。
	var rds *redis.Client
	var err error
	if !c.RedisDisabled {
		rds, err = redis.NewClient(c.Redis)
		if err != nil {
			log.Fatalf("fatal: failed to init redis: %v (fail-closed)", err)
		}
	}

	return &ServiceContext{
		Config:       c,
		Redis:        rds,
		SysUsers:     apps.NewSysUsersClient(conn.Conn()),
		SysRoles:     apps.NewSysRolesClient(conn.Conn()),
		SysDepts:     apps.NewSysDeptsClient(conn.Conn()),
		SysMenus:     apps.NewSysMenusClient(conn.Conn()),
		SysDicts:     apps.NewSysDictsClient(conn.Conn()),
		SysDictDatas: apps.NewSysDictDatasClient(conn.Conn()),
		SysTenants:   apps.NewSysTenantsClient(conn.Conn()),
		SysPackages:  apps.NewSysPackagesClient(conn.Conn()),
		SysApis:      apps.NewSysApisClient(conn.Conn()),
		SysLogs:      apps.NewSysLogsClient(conn.Conn()),
		SysInit:      apps.NewSysInitClient(conn.Conn()),
		Enforcer:     enf,
	}
}
