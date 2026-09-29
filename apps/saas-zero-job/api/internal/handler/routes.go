package handler

import (
	"github.com/saas-zero/saas-zero-job/api/internal/svc"
	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{
		{Method: "GET", Path: "/system/job/handlers", Handler: GetJobHandlersHandler(serverCtx)},
		{Method: "POST", Path: "/system/job/create", Handler: CreateJobHandler(serverCtx)},
		{Method: "POST", Path: "/system/job/update", Handler: UpdateJobHandler(serverCtx)},
		{Method: "POST", Path: "/system/job/delete", Handler: DeleteJobHandler(serverCtx)},
		{Method: "GET", Path: "/system/job/list", Handler: GetJobListHandler(serverCtx)},
		{Method: "GET", Path: "/system/job/detail", Handler: GetJobDetailHandler(serverCtx)},
		{Method: "POST", Path: "/system/job/start", Handler: StartJobHandler(serverCtx)},
		{Method: "POST", Path: "/system/job/pause", Handler: PauseJobHandler(serverCtx)},
		{Method: "POST", Path: "/system/job/runOnce", Handler: RunJobOnceHandler(serverCtx)},
		{Method: "POST", Path: "/system/job/log/clean", Handler: CleanJobLogHandler(serverCtx)},
		{Method: "GET", Path: "/system/job/log/list", Handler: GetJobLogListHandler(serverCtx)},
	})
}
