// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/saas-zero/saas-zero-common/pkg/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	JwtSecret   string     `json:"jwtSecret"`
	JwtExpire   int64      `json:"jwtExpire"`
	Redis       redis.Conf `json:"redis"`
	BaseDataRpc zrpc.RpcClientConf
}
