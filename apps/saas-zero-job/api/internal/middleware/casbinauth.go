package middleware

import (
	"net/http"
	"strconv"

	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"

	casbinapi "github.com/casbin/casbin/v2"
)

// CasbinAuth returns HTTP middleware enforcing Casbin Domain RBAC.
// RoleCodes are read from JWT claims (set by JwtAuth middleware via context),
// then checked against Casbin policy for each role.
// fail-closed：disabled=false 时 enforcer 为 nil（初始化失败）一律 503 拒绝。
func CasbinAuth(enf *casbinapi.SyncedEnforcer, disabled bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if disabled {
				next(w, r)
				return
			}
			if enf == nil {
				writeJSON(w, http.StatusInternalServerError, errno.AuthServiceUnavailable)
				return
			}
			if r.URL.Path == "" || r.URL.Path[0] != '/' {
				next(w, r)
				return
			}
			if len(r.URL.Path) >= 6 && r.URL.Path[:6] == "/init/" {
				next(w, r)
				return
			}
			// handler 列表接口：只读，返回当前注册的所有 handler code，不参与策略校验
			if r.URL.Path == "/system/job/handlers" {
				next(w, r)
				return
			}
			tenantId := mixins.GetCurrentTenantId(r.Context())
			roleCodes := GetRoleCodes(r.Context())
			if len(roleCodes) == 0 {
				writeJSON(w, http.StatusForbidden, errno.NoRoles)
				return
			}
			path := r.URL.Path
			method := r.Method
			dom := strconv.FormatInt(tenantId, 10)
			allowed := false
			for _, roleCode := range roleCodes {
				ok, err := enf.Enforce(roleCode, dom, path, method)
				if err != nil {
					logx.Errorf("Casbin enforce error: role=%s, dom=%s, path=%s, method=%s, err=%v",
						roleCode, dom, path, method, err)
					writeJSON(w, http.StatusInternalServerError, errno.AuthServiceUnavailable)
					return
				}
				if ok {
					allowed = true
					break
				}
			}
			if !allowed {
				writeJSON(w, http.StatusForbidden, errno.ForbiddenOperation)
				return
			}
			next(w, r)
		}
	}
}
