package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/saas-zero/saas-zero-job/api/internal/config"
	"github.com/saas-zero/saas-zero-job/api/internal/handler"
	"github.com/saas-zero/saas-zero-job/api/internal/middleware"
	"github.com/saas-zero/saas-zero-job/api/internal/svc"

	"github.com/saas-zero/saas-zero-common/pkg/envconf"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var configFile = flag.String("f", "etc/jobservice.yaml", "the config file")

// jobErrHandler 在通用 errno 处理器之上，把 job RPC 返回的 gRPC 状态错误还原为
// 业务码+消息（RPC 侧 errnoInterceptor 做了业务码 → 状态码的映射）。
// 不做这层还原的话，RPC 抛出的业务校验错误会在 HTTP 层退化成
// 500 “内部服务器错误”，前端只能看到一个没有原因的弹框。
func jobErrHandler(err error) (int, any) {
	if st, ok := status.FromError(err); ok {
		var code int
		switch st.Code() {
		case codes.InvalidArgument:
			code = errno.InvalidParam.Code
		case codes.PermissionDenied:
			code = errno.Forbidden.Code
		case codes.Unauthenticated:
			code = errno.Unauthorized.Code
		}
		if code != 0 {
			return code, map[string]any{"code": code, "msg": st.Message()}
		}
	}
	return errno.ErrHandler(err)
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 环境变量覆盖（生产用），无则使用 YAML 明文（本地调试）
	c.JwtSecret = envconf.String("JWT_SECRET", c.JwtSecret)
	c.CasbinPostgres.DataSource = envconf.String("CASBIN_POSTGRES_DSN", c.CasbinPostgres.DataSource)
	c.Redis.Host = envconf.String("REDIS_HOST", c.Redis.Host)
	c.Redis.Pass = envconf.String("REDIS_PASS", c.Redis.Pass)
	if db := os.Getenv("REDIS_DB"); db != "" {
		if n, err := strconv.Atoi(db); err == nil {
			c.Redis.DB = n
		}
	}

	// 统一错误响应（code 取自 common/errno；RPC 业务错误还原真实 msg）
	httpx.SetErrorHandler(jobErrHandler)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	server.Use(middleware.JwtAuth(c.JwtSecret, ctx.Redis, c.RedisDisabled))
	server.Use(middleware.CasbinAuth(ctx.Enforcer, c.CasbinDisabled))

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
