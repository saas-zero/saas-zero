package handler

import (
	"net/http"

	"github.com/saas-zero/saas-zero-job/api/internal/config"
	"github.com/saas-zero/saas-zero-job/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type handlerItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func GetJobHandlersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items := make([]handlerItem, 0, len(svcCtx.Config.Handlers))
		for _, h := range svcCtx.Config.Handlers {
			items = append(items, handlerItem{Code: h.Code, Name: h.Name})
		}
		httpx.OkJsonCtx(r.Context(), w, items)
	}
}

// convertHandlers 将配置中的 HandlerOption 列表转为 handlerItem 列表
func convertHandlers(handlers []config.HandlerOption) []handlerItem {
	items := make([]handlerItem, 0, len(handlers))
	for _, h := range handlers {
		items = append(items, handlerItem{Code: h.Code, Name: h.Name})
	}
	return items
}
