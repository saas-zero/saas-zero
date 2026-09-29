// Code scaffolded by goctl. Safe to edit.

package handler

import (
	"net/http"

	"github.com/saas-zero/saas-zero-auth/api/internal/logic"
	"github.com/saas-zero/saas-zero-auth/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func OauthPermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := logic.WithToken(r.Context(), logic.ExtractBearerToken(r.Header.Get("Authorization")))
		l := logic.NewOauthPermissionsLogic(ctx, svcCtx)
		resp, err := l.OauthPermissions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
