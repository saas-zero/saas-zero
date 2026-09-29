// Code scaffolded by goctl. Safe to edit.

package handler

import (
	"net/http"

	"github.com/saas-zero/saas-zero-auth/api/internal/logic"
	"github.com/saas-zero/saas-zero-auth/api/internal/svc"
	"github.com/saas-zero/saas-zero-auth/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func PasswordResetHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PasswordResetReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		ctx := logic.WithToken(r.Context(), logic.ExtractBearerToken(r.Header.Get("Authorization")))
		l := logic.NewPasswordResetLogic(ctx, svcCtx)
		resp, err := l.PasswordReset(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
