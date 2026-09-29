package handler

import (
	"net/http"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/zeromicro/go-zero/rest"
)

func RegisterInitRoutes(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/init/all",
				Handler: InitAllHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/init/package/create",
				Handler: InitCreatePackageHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/init/tenant/create",
				Handler: InitCreateTenantHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/init/user/create",
				Handler: InitCreateUserHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/init/role/create",
				Handler: InitCreateRoleHandler(serverCtx),
			},
		},
	)
}
