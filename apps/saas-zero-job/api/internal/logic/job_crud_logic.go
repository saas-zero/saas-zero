package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-job/api/internal/svc"
	"github.com/saas-zero/saas-zero-job/api/internal/types"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetJobListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetJobListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetJobListLogic {
	return &GetJobListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetJobListLogic) GetJobList(req *types.JobPageReq) (*types.JobPageResp, error) {
	rpcReq := &apps.JobListReq{
		Name:     req.Name,
		Group:    req.Group,
		Handler:  req.Handler,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	resp, err := l.svcCtx.JobRpc.GetJobList(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	list := make([]*types.Job, 0, len(resp.GetList()))
	for _, p := range resp.GetList() {
		list = append(list, protoToJob(p))
	}
	return &types.JobPageResp{List: list, Total: resp.GetTotal()}, nil
}

type GetJobDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetJobDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetJobDetailLogic {
	return &GetJobDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetJobDetailLogic) GetJobDetail(id string) (*types.Job, error) {
	resp, err := l.svcCtx.JobRpc.GetJobById(l.ctx, &apps.IdReq{Id: id})
	if err != nil {
		return nil, err
	}
	return protoToJob(resp), nil
}

type UpdateJobLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateJobLogic {
	return &UpdateJobLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateJobLogic) UpdateJob(req *types.Job) (*types.BaseResp, error) {
	resp, err := l.svcCtx.JobRpc.UpdateJob(l.ctx, jobToProto(req))
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}

type DeleteJobLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteJobLogic {
	return &DeleteJobLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteJobLogic) DeleteJob(ids []string) (*types.BaseResp, error) {
	resp, err := l.svcCtx.JobRpc.DeleteJob(l.ctx, &apps.IdsReq{Ids: ids})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}

type StartJobLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStartJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartJobLogic {
	return &StartJobLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StartJobLogic) StartJob(id string) (*types.BaseResp, error) {
	resp, err := l.svcCtx.JobRpc.StartJob(l.ctx, &apps.IdReq{Id: id})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}

type PauseJobLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPauseJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PauseJobLogic {
	return &PauseJobLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PauseJobLogic) PauseJob(id string) (*types.BaseResp, error) {
	resp, err := l.svcCtx.JobRpc.PauseJob(l.ctx, &apps.IdReq{Id: id})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}

type RunJobOnceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRunJobOnceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunJobOnceLogic {
	return &RunJobOnceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RunJobOnceLogic) RunJobOnce(id string) (*types.BaseResp, error) {
	resp, err := l.svcCtx.JobRpc.RunJobOnce(l.ctx, &apps.IdReq{Id: id})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}

type GetJobLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetJobLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetJobLogListLogic {
	return &GetJobLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetJobLogListLogic) GetJobLogList(req *types.JobLogPageReq) (*types.JobLogPageResp, error) {
	rpcReq := &apps.JobLogListReq{
		JobId:     req.JobId,
		JobName:   req.JobName,
		Status:    req.Status,
		BeginTime: req.BeginTime,
		EndTime:   req.EndTime,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}
	resp, err := l.svcCtx.JobRpc.GetJobLogList(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	list := make([]*types.JobLog, 0, len(resp.GetList()))
	for _, p := range resp.GetList() {
		list = append(list, protoToJobLog(p))
	}
	return &types.JobLogPageResp{List: list, Total: resp.GetTotal()}, nil
}

type CleanJobLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCleanJobLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CleanJobLogLogic {
	return &CleanJobLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CleanJobLogLogic) CleanJobLog(req *types.CleanJobLogReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.JobRpc.CleanJobLog(l.ctx, &apps.CleanJobLogReq{
		JobId:    req.JobId,
		KeepDays: req.KeepDays,
	})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}
