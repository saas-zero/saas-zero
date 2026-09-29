package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/jwt"
	"github.com/saas-zero/saas-zero-common/pkg/redis"
)

type ctxKey string

const roleCodesKey ctxKey = "role_codes"

// GetRoleCodes 从 context 取角色码列表
func GetRoleCodes(ctx context.Context) []string {
	if v, ok := ctx.Value(roleCodesKey).([]string); ok {
		return v
	}
	return nil
}

// writeJSON 以标准 JSON 输出错误响应
func writeJSON(w http.ResponseWriter, status int, e *errno.Errno) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write([]byte(e.JSON()))
}

// JwtAuth 校验 JWT 签名、Redis 会话与 tokenVersion；解析后注入用户上下文。
// rds 为 nil 时仅当 skipSessionChecks=true（显式本地开发配置）放行。
func JwtAuth(secret string, rds *redis.Client, skipSessionChecks bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				writeJSON(w, http.StatusUnauthorized, errno.MissingAuthHeader)
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				writeJSON(w, http.StatusUnauthorized, errno.InvalidAuthHeader)
				return
			}
			claims, err := jwt.Parse(parts[1], secret)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, errno.InvalidToken)
				return
			}
			if rds == nil {
				if !skipSessionChecks {
					writeJSON(w, http.StatusInternalServerError, errno.AuthServiceUnavailable)
					return
				}
			} else {
				if claims.ID != "" {
					exists, err := rds.Exists(fmt.Sprintf("token:%s", claims.ID))
					if err != nil || !exists {
						writeJSON(w, http.StatusUnauthorized, errno.TokenInvalidated)
						return
					}
				}
				tv, err := rds.Get(fmt.Sprintf("token_version:%d", claims.UserId))
				if err != nil {
					writeJSON(w, http.StatusUnauthorized, errno.TokenInvalidated)
					return
				}
				if tv != fmt.Sprintf("%d", claims.TokenVersion) {
					writeJSON(w, http.StatusUnauthorized, errno.TokenInvalidated)
					return
				}
			}

			// 注入用户上下文，出站时附到 gRPC metadata（审计字段）
			ctx := mixins.SetCurrentUserId(r.Context(), claims.UserId)
			ctx = mixins.SetCurrentUserName(ctx, claims.UserName)
			ctx = mixins.SetCurrentTenantId(ctx, claims.TenantId)
			ctx = context.WithValue(ctx, roleCodesKey, claims.RoleCodes)
			next(w, r.WithContext(ctx))
		}
	}
}
