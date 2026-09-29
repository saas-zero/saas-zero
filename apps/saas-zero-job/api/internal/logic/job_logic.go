package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-job/api/internal/svc"
	"github.com/saas-zero/saas-zero-job/api/internal/types"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateJobLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateJobLogic {
	return &CreateJobLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateJobLogic) CreateJob(req *types.Job) (*types.BaseResp, error) {
	resp, err := l.svcCtx.JobRpc.CreateJob(l.ctx, jobToProto(req))
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.Msg}, nil
}

func jobToProto(req *types.Job) *apps.Job {
	p := &apps.Job{
		Name:           req.Name,
		Group:          req.Group,
		Handler:        req.Handler,
		Params:         req.Params,
		CronExpression: req.CronExpression,
		TimeZone:       req.TimeZone,
		MisfirePolicy:  req.MisfirePolicy,
		Concurrent:     req.Concurrent,
		Timeout:        req.Timeout,
		MaxRetry:       req.MaxRetry,
		RetryInterval:  req.RetryInterval,
		Status:         req.Status,
		Remark:         req.Remark,
	}
	if req.Id != "" {
		p.Id = req.Id
	}
	return p
}

func protoToJob(p *apps.Job) *types.Job {
	if p == nil {
		return nil
	}
	return &types.Job{
		Id:             p.GetId(),
		CreatedAt:      p.GetCreatedAt(),
		UpdatedAt:      p.GetUpdatedAt(),
		Name:           p.GetName(),
		Group:          p.GetGroup(),
		Handler:        p.GetHandler(),
		Params:         p.GetParams(),
		CronExpression: p.GetCronExpression(),
		TimeZone:       p.GetTimeZone(),
		MisfirePolicy:  p.GetMisfirePolicy(),
		Concurrent:     p.GetConcurrent(),
		Timeout:        p.GetTimeout(),
		MaxRetry:       p.GetMaxRetry(),
		RetryInterval:  p.GetRetryInterval(),
		Status:         p.GetStatus(),
		Remark:         p.GetRemark(),
		NextRunAt:      p.GetNextRunAt(),
		LastRunAt:      p.GetLastRunAt(),
		LastStatus:     p.GetLastStatus(),
		LastDuration:   p.GetLastDuration(),
		LastError:      p.GetLastError(),
	}
}

func protoToJobLog(p *apps.JobLog) *types.JobLog {
	if p == nil {
		return nil
	}
	return &types.JobLog{
		Id:            p.GetId(),
		CreatedAt:     p.GetCreatedAt(),
		JobId:         p.GetJobId(),
		JobName:       p.GetJobName(),
		JobGroup:      p.GetJobGroup(),
		Handler:       p.GetHandler(),
		TriggerType:   p.GetTriggerType(),
		ExecNode:      p.GetExecNode(),
		Status:        p.GetStatus(),
		Attempt:       p.GetAttempt(),
		Message:       p.GetMessage(),
		ExceptionInfo: p.GetExceptionInfo(),
		Duration:      p.GetDuration(),
	}
}
