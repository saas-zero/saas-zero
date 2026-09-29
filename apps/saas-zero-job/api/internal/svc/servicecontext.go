package svc

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"time"

	_ "github.com/lib/pq"

	casbinapi "github.com/casbin/casbin/v2"
	commcasbin "github.com/saas-zero/saas-zero-common/pkg/casbin"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/redis"
	"github.com/saas-zero/saas-zero-job/api/internal/config"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServiceContext struct {
	Config   config.Config
	Redis    *redis.Client
	JobRpc   apps.SysJobsClient
	Enforcer *casbinapi.SyncedEnforcer
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Redis 客户端初始化：fail-closed
	var rds *redis.Client
	if !c.RedisDisabled {
		var err error
		rds, err = redis.NewClient(c.Redis)
		if err != nil {
			log.Fatalf("failed initializing redis: %v (fail-closed)", err)
		}
	}

	// Casbin enforcer 初始化：fail-closed
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

	conn := zrpc.MustNewClient(c.JobRpc, zrpc.WithUnaryClientInterceptor(authClientInterceptor))
	return &ServiceContext{
		Config:   c,
		Redis:    rds,
		JobRpc:   apps.NewSysJobsClient(conn.Conn()),
		Enforcer: enf,
	}
}

// authClientInterceptor 把当前用户上下文附加到 gRPC 出站 metadata，
// job RPC 侧 authInterceptor 读回并注入审计字段。
func authClientInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	uid := mixins.GetCurrentUserId(ctx)
	uname := mixins.GetCurrentUserName(ctx)
	tid := mixins.GetCurrentTenantId(ctx)

	if uid > 0 || uname != "" {
		md := metadata.Pairs(
			"x-user-id", strconv.FormatInt(uid, 10),
			"x-user-name", uname,
			"x-tenant-id", strconv.FormatInt(tid, 10),
		)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}
