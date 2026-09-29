// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/saas-zero/saas-zero-auth/api/internal/config"
	"github.com/saas-zero/saas-zero-auth/api/internal/handler"
	"github.com/saas-zero/saas-zero-auth/api/internal/middleware"
	"github.com/saas-zero/saas-zero-auth/api/internal/svc"

	"github.com/saas-zero/saas-zero-common/pkg/envconf"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/authApis.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 环境变量覆盖（生产用），无则使用 YAML 明文（本地调试）
	c.JwtSecret = envconf.String("JWT_SECRET", c.JwtSecret)
	c.Redis.Host = envconf.String("REDIS_HOST", c.Redis.Host)
	c.Redis.Pass = envconf.String("REDIS_PASS", c.Redis.Pass)
	if db := os.Getenv("REDIS_DB"); db != "" {
		if n, err := strconv.Atoi(db); err == nil {
			c.Redis.DB = n
		}
	}

	// 统一错误响应：code 全部取自 common/errno（见 errno.ErrHandler）
	httpx.SetErrorHandler(errno.ErrHandler)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	server.Use(middleware.ClientIP())
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
