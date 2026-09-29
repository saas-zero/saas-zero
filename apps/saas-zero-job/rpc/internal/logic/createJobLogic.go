package logic

import (
	"context"
	"strconv"

	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-job/ent/sysjob"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateJobLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateJobLogic {
	return &CreateJobLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateJob 创建任务
func (l *CreateJobLogic) CreateJob(in *apps.Job) (*apps.EmptyResp, error) {
	if in == nil {
		return nil, errInvalidParam("job is required")
	}
	if in.Name == "" {
		return nil, errInvalidParam("name is required")
	}
	if in.Handler == "" {
		return nil, errInvalidParam("handler is required")
	}
	if !handlerRegistered(l.svcCtx.Scheduler.Registry(), in.Handler) {
		return nil, errInvalidParam("handler %q not registered", in.Handler)
	}
	if in.CronExpression == "" {
		return nil, errInvalidParam("cron_expression is required")
	}
	if err := validateCron(in.CronExpression); err != nil {
		return nil, errInvalidParam("invalid cron expression: %v", err)
	}
	if err := validateParams(in.Params); err != nil {
		return nil, err
	}

	group := in.Group
	if group == "" {
		group = "default"
	}
	timeZone := in.TimeZone
	if timeZone == "" {
		timeZone = "Asia/Shanghai"
	}
	misfire := in.MisfirePolicy
	if misfire == "" {
		misfire = "fire_once"
	}
	timeout := in.Timeout
	if timeout <= 0 {
		timeout = 30
	}
	maxRetry := in.MaxRetry
	if maxRetry < 0 {
		maxRetry = 0
	}
	retryInterval := in.RetryInterval
	if retryInterval <= 0 {
		retryInterval = 10
	}

	ctx := systemCtx(l.ctx)
	ctx = mixins.SetCurrentTenantId(ctx, 0) // job 平台级，无租户隔离

	j, err := l.svcCtx.DB.SysJob.Create().
		SetName(in.Name).
		SetGroup(group).
		SetHandler(in.Handler).
		SetParams(in.Params).
		SetCronExpression(in.CronExpression).
		SetTimeZone(timeZone).
		SetMisfirePolicy(sysjob.MisfirePolicy(misfire)).
		SetConcurrent(in.Concurrent).
		SetTimeout(timeout).
		SetMaxRetry(maxRetry).
		SetRetryInterval(retryInterval).
		SetStatus(sysjob.StatusActive).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		l.Logger.Errorf("create job failed: %v", err)
		return nil, err
	}

	l.svcCtx.Scheduler.SyncNow()
	l.Logger.Infof("job created: id=%d name=%s handler=%s", j.ID, j.Name, j.Handler)
	return &apps.EmptyResp{Code: 0, Msg: strconv.FormatInt(j.ID, 10)}, nil
}
