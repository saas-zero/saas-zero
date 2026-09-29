package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync/atomic"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/lib/pq"
	"github.com/saas-zero/saas-zero-common/pkg/redis"
	"github.com/saas-zero/saas-zero-job/ent"
	_ "github.com/saas-zero/saas-zero-job/ent/runtime"
	"github.com/saas-zero/saas-zero-job/rpc/internal/config"
	"github.com/saas-zero/saas-zero-job/rpc/internal/scheduler"
	"github.com/saas-zero/saas-zero-job/rpc/internal/task"
	"github.com/zeromicro/go-zero/core/logx"
)

const redisKeyHandlers = "job:handlers"

// ServiceContext job 服务上下文
type ServiceContext struct {
	Config    config.Config
	DB        *ent.Client
	Redis     *redis.Client
	Scheduler *scheduler.Scheduler
}

func NewServiceContext(c config.Config) *ServiceContext {
	client, err := ent.Open(dialect.Postgres, c.Postgres.DataSource)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v (fail-closed)", err)
	}
	if c.Postgres.Debug {
		client = client.Debug()
	}
	// 独立数据库：job 表只在本库创建，不影响 saas_zero_kun
	if err := client.Schema.Create(context.Background(), schema.WithDropIndex(true)); err != nil {
		log.Fatalf("failed creating schema resources: %v (fail-closed)", err)
	}

	rds, err := redis.NewClient(c.CacheRedis)
	if err != nil {
		log.Fatalf("failed initializing redis: %v (fail-closed)", err)
	}

	node := schedulerNode(c.Scheduler.Node)
	sch := scheduler.NewScheduler(scheduler.Options{
		DB:               client,
		Redis:            rds,
		SyncIntervalSec:  c.Scheduler.SyncInterval,
		LogRetentionDays: c.Scheduler.LogRetentionDays,
		CleanupHour:      c.Scheduler.CleanupHour,
		Node:             node,
	})

	// 注册内置任务处理器（必须在调度器 Start 之前，否则 Start 后首个触发周期可能命中未注册 handler）
	if err := task.RegisterAll(sch.Registry()); err != nil {
		log.Fatalf("failed registering task handlers: %v (fail-closed)", err)
	}

	// 将已注册的 handler 列表写入 Redis，供 job API HTTP 层读取
	if err := publishHandlers(rds, sch.Registry()); err != nil {
		logx.Errorf("publish handlers to redis: %v", err)
	}

	// 启动调度器：失败即中止服务（fail-closed），避免分布式环境下误触发
	if c.Scheduler.Enabled {
		if err := sch.Start(); err != nil {
			log.Fatalf("failed starting scheduler: %v (fail-closed)", err)
		}
		logx.Infof("scheduler started with node=%s", node)
	}

	return &ServiceContext{
		Config:    c,
		DB:        client,
		Redis:     rds,
		Scheduler: sch,
	}
}

// publishHandlers 将已注册的 handler code 列表写入 Redis（JSON 数组），供 HTTP 层读取。
func publishHandlers(rds *redis.Client, reg *scheduler.Registry) error {
	keys := reg.Keys()
	sort.Strings(keys)
	type handlerItem struct {
		Code string `json:"code"`
	}
	items := make([]handlerItem, 0, len(keys))
	for _, k := range keys {
		items = append(items, handlerItem{Code: k})
	}
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return rds.Setex(redisKeyHandlers, string(data), 86400)
}

var schedulerNodeSeq atomic.Int64

// schedulerNode 生成实例标识：优先配置，否则 hostname:pid:seq
// 同机短生命周期多次启动时（测试/开发）用 seq 区分。
func schedulerNode(configured string) string {
	if configured != "" {
		return configured
	}
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unknown"
	}
	return fmt.Sprintf("%s:%d:%d", host, os.Getpid(), schedulerNodeSeq.Add(1))
}
