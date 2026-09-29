package config

import (
	"github.com/saas-zero/saas-zero-common/pkg/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type PostgresConfig struct {
	DataSource string
	// Debug 开启后 ent 打印全部 SQL 及参数（含敏感值），仅限本地排障。
	Debug bool `json:",optional"`
}

type Config struct {
	zrpc.RpcServerConf
	Postgres   PostgresConfig
	CacheRedis redis.Conf `json:"cacheRedis"`
	// Scheduler 调度器配置
	Scheduler SchedulerConfig `json:"scheduler"`
}

type SchedulerConfig struct {
	// Enabled 是否开启调度器（false 时仅提供 CRUD，不触发任务）
	Enabled bool `json:"enabled"`
	// SyncInterval 数据库对齐轮询间隔（秒），默认 30
	SyncInterval int `json:"syncInterval,optional,default=30"`
	// LogRetentionDays 任务日志保留天数，默认 30
	LogRetentionDays int `json:"logRetentionDays,optional,default=30"`
	// CleanupHour 每日日志清理执行的小时（0-23），默认 3
	CleanupHour int `json:"cleanupHour,optional,default=3"`
	// Node 执行实例标识（多实例排查，默认取主机名+进程号）
	Node string `json:"node,optional"`
}
