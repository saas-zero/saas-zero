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
}
