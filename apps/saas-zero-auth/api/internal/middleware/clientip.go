package middleware

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type ctxKey string

const clientIPKey ctxKey = "client_ip"

// ClientIP 提取客户端 IP 并注入 context，供登录逻辑记录登录 IP / 登录日志使用。
func ClientIP() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), clientIPKey, httpx.GetRemoteAddr(r))
			next(w, r.WithContext(ctx))
		}
	}
}

// GetClientIP 从 context 读取客户端 IP，未设置时返回空串。
func GetClientIP(ctx context.Context) string {
	if v, ok := ctx.Value(clientIPKey).(string); ok {
		return v
	}
	return ""
}
