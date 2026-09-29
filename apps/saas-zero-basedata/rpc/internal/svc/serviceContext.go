package svc

import (
	"context"
	"database/sql"
	"log"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	casbinapi "github.com/casbin/casbin/v2"
	_ "github.com/lib/pq"
	"github.com/saas-zero/saas-zero-basedata/ent"
	_ "github.com/saas-zero/saas-zero-basedata/ent/runtime"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/config"
	commcasbin "github.com/saas-zero/saas-zero-common/pkg/casbin"
	"github.com/saas-zero/saas-zero-common/pkg/redis"
)

type ServiceContext struct {
	Config config.Config
	DB     *ent.Client
	// Redis 在 RPC 侧只用于递增 token_version（会话失效）。
	// 声明为窄接口便于测试注入失败场景，生产由 *redis.Client 实现。
	Redis    redis.TokenVersionStore
	Enforcer *casbinapi.SyncedEnforcer
}

func NewServiceContext(c config.Config) *ServiceContext {
	client, err := ent.Open(dialect.Postgres, c.Postgres.DataSource)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	// Debug 会打印全部 SQL 及绑定参数（含密码哈希、手机号等敏感值），
	// 仅允许本地排障通过配置显式开启，生产必须保持关闭。
	if c.Postgres.Debug {
		client = client.Debug()
	}
	// WithDropIndex: 删除 schema 中已不存在的索引（如旧的字段级唯一索引），
	// 确保条件唯一索引迁移后旧索引不会残留导致唯一约束仍生效。
	if err := client.Schema.Create(context.Background(), schema.WithDropIndex(true)); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	if err := SeedSystemDicts(context.Background(), client); err != nil {
		log.Fatalf("failed seeding system dictionaries: %v", err)
	}
	casbinDb, err := sql.Open("postgres", c.Postgres.DataSource)
	if err != nil {
		log.Fatalf("failed opening casbin db: %v (fail-closed)", err)
	}
	enf, err := commcasbin.NewEnforcer(casbinDb, "casbin_rule")
	if err != nil {
		log.Fatalf("failed initializing casbin: %v (fail-closed)", err)
	}
	rds, err := redis.NewClient(c.CacheRedis)
	if err != nil {
		log.Fatalf("failed initializing redis: %v (fail-closed)", err)
	}
	return &ServiceContext{
		Config:   c,
		DB:       client,
		Redis:    rds,
		Enforcer: enf,
	}
}
