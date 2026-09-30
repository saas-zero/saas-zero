package config

import (
	"github.com/saas-zero/saas-zero-common/pkg/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type CasbinPostgresConfig struct {
	DataSource string `json:"dataSource"`
}

type Config struct {
	rest.RestConf
	JwtSecret      string               `json:"jwtSecret"`
	Redis          redis.Conf           `json:"redis"`
	CasbinPostgres CasbinPostgresConfig `json:"casbinPostgres"`
	JobRpc         zrpc.RpcClientConf   `json:"jobRpc"`
	CasbinDisabled bool                 `json:"casbinDisabled,optional"`
	RedisDisabled  bool                 `json:"redisDisabled,optional"`
}
