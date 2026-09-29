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
	// Handlers 已注册的处理器列表（code），供前端下拉选择。
	// 新增 handler 时在此处添加即可，与 rpc/internal/task/handlers.go 保持一致。
	Handlers       []HandlerOption `json:"handlers"`
	CasbinDisabled bool            `json:"casbinDisabled,optional"`
	RedisDisabled  bool            `json:"redisDisabled,optional"`
}

type HandlerOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
