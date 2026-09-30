package handler

import (
	"net/http"

	"github.com/saas-zero/saas-zero-job/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/protobuf/types/known/emptypb"
)

type handlerItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// GetJobHandlersHandler 返回可选的任务处理器列表。
// 数据源是 job RPC 进程内的代码注册表（GetRegisteredHandlers），不是配置文件副本，
// 因此新增/删除 handler 时无需同步 yaml，管理端列表也不会与实际可执行项漂移。
func GetJobHandlersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := svcCtx.JobRpc.GetRegisteredHandlers(r.Context(), &emptypb.Empty{})
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		list := resp.GetList()
		items := make([]handlerItem, 0, len(list))
		for _, h := range list {
			items = append(items, handlerItem{Code: h.GetCode(), Name: h.GetName()})
		}
		httpx.OkJsonCtx(r.Context(), w, items)
	}
}
