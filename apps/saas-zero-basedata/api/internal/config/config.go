// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

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
	Basedata       zrpc.RpcClientConf
	// CasbinDisabled 仅允许本地开发通过显式配置关闭 Casbin 校验。
	// 默认为 false（依赖 go-zero 默认值），生产环境必须保留校验。
	CasbinDisabled bool `json:"casbinDisabled,optional"`
	// RedisDisabled 仅供本地开发跳过 Redis 会话校验使用，生产不得开启。
	RedisDisabled bool `json:"redisDisabled,optional"`
}
