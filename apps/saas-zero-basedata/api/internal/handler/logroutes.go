package handler

import (
	"net/http"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/zeromicro/go-zero/rest"
)

func RegisterLogRoutes(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/system/log/loginLog/list",
				Handler: GetLoginLogListHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/system/log/operationLog/list",
				Handler: GetOperationLogListHandler(serverCtx),
			},
		},
	)
}
