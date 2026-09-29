package server

import (
	"context"

	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/logic"
	"github.com/saas-zero/saas-zero-job/rpc/internal/svc"
)

type SysJobsServer struct {
	apps.UnimplementedSysJobsServer
	svcCtx *svc.ServiceContext
}

func NewSysJobsServer(svcCtx *svc.ServiceContext) *SysJobsServer {
	return &SysJobsServer{svcCtx: svcCtx}
}

func (s *SysJobsServer) GetJobById(ctx context.Context, in *apps.IdReq) (*apps.Job, error) {
	return logic.NewGetJobByIdLogic(ctx, s.svcCtx).GetJobById(in)
}

func (s *SysJobsServer) GetJobList(ctx context.Context, in *apps.JobListReq) (*apps.JobListResp, error) {
	return logic.NewGetJobListLogic(ctx, s.svcCtx).GetJobList(in)
}

func (s *SysJobsServer) CreateJob(ctx context.Context, in *apps.Job) (*apps.EmptyResp, error) {
	return logic.NewCreateJobLogic(ctx, s.svcCtx).CreateJob(in)
}

func (s *SysJobsServer) UpdateJob(ctx context.Context, in *apps.Job) (*apps.EmptyResp, error) {
	return logic.NewUpdateJobLogic(ctx, s.svcCtx).UpdateJob(in)
}

func (s *SysJobsServer) DeleteJob(ctx context.Context, in *apps.IdsReq) (*apps.EmptyResp, error) {
	return logic.NewDeleteJobLogic(ctx, s.svcCtx).DeleteJob(in)
}

func (s *SysJobsServer) StartJob(ctx context.Context, in *apps.IdReq) (*apps.EmptyResp, error) {
	return logic.NewStartJobLogic(ctx, s.svcCtx).StartJob(in)
}

func (s *SysJobsServer) PauseJob(ctx context.Context, in *apps.IdReq) (*apps.EmptyResp, error) {
	return logic.NewPauseJobLogic(ctx, s.svcCtx).PauseJob(in)
}

func (s *SysJobsServer) RunJobOnce(ctx context.Context, in *apps.IdReq) (*apps.EmptyResp, error) {
	return logic.NewRunJobOnceLogic(ctx, s.svcCtx).RunJobOnce(in)
}

func (s *SysJobsServer) CleanJobLog(ctx context.Context, in *apps.CleanJobLogReq) (*apps.EmptyResp, error) {
	return logic.NewCleanJobLogLogic(ctx, s.svcCtx).CleanJobLog(in)
}

func (s *SysJobsServer) GetJobLogList(ctx context.Context, in *apps.JobLogListReq) (*apps.JobLogListResp, error) {
	return logic.NewGetJobLogListLogic(ctx, s.svcCtx).GetJobLogList(in)
}
