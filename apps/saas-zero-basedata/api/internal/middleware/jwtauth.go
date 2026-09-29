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

func GetRoleCodes(ctx context.Context) []string {
	if v, ok := ctx.Value(roleCodesKey).([]string); ok {
		return v
	}
	return nil
}

// writeJSON 以标准 JSON 形式输出错误响应。
// 不使用 http.Error（其 Content-Type 为 text/plain，axios 不会解析为对象，
// 导致前端无法识别 code 字段而无法自动登出）。
func writeJSON(w http.ResponseWriter, status int, e *errno.Errno) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write([]byte(e.JSON()))
}

// JwtAuth 校验 JWT 签名、有效期、Redis JTI 存在性与 tokenVersion。
// rds 为 nil 时：
//   - skipSessionChecks=true（显式本地开发配置 redisDisabled）允许跳过会话校验；
//   - 否则 fail-closed，拒绝请求（正常生产启动时 Redis 初始化失败会直接终止进程，
//     不会走到这里）。
func JwtAuth(secret string, rds *redis.Client, skipSessionChecks bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/init/") {
				next(w, r)
				return
			}
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
				// 禁用 Redis 会话校验仅允许在显式本地开发配置下使用
				if !skipSessionChecks {
					writeJSON(w, http.StatusInternalServerError, errno.AuthServiceUnavailable)
					return
				}
			} else {
				// Verify token exists in Redis (allows remote invalidation)
				if claims.ID != "" {
					exists, err := rds.Exists(fmt.Sprintf("token:%s", claims.ID))
					if err != nil || !exists {
						writeJSON(w, http.StatusUnauthorized, errno.TokenInvalidated)
						return
					}
				}
				// Verify token version matches Redis (invalidates on role/permission/password changes)
				// 始终校验版本：即使 TokenVersion=0（多会话共存的首个版本）也参与比较，
				// 否则权限/密码变更后 version=0 的旧 token 会绕过踢出逻辑。
				tv, err := rds.Get(fmt.Sprintf("token_version:%d", claims.UserId))
				if err != nil || tv == "" || tv != fmt.Sprintf("%d", claims.TokenVersion) {
					writeJSON(w, http.StatusUnauthorized, errno.TokenVersionMismatch)
					return
				}
			}
			ctx := r.Context()
			// Inject JWT claims into request context for downstream use.
			// - UserId/UserName/TenantId are used by ent mixin hooks for audit fields.
			// - RoleCodes are consumed by CasbinAuth middleware for authorization checks,
			//   eliminating a gRPC call to GetUserRoleCodes on every request.
			ctx = mixins.SetCurrentUserId(ctx, claims.UserId)
			ctx = mixins.SetCurrentUserName(ctx, claims.UserName)
			ctx = mixins.SetCurrentTenantId(ctx, claims.TenantId)
			ctx = context.WithValue(ctx, roleCodesKey, claims.RoleCodes)
			next(w, r.WithContext(ctx))
		}
	}
}
