// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/saas-zero/saas-zero-basedata/api/internal/logic"
	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetMenuTreeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetMenuTreeLogic(r.Context(), svcCtx)
		resp, err := l.GetMenuTree()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
