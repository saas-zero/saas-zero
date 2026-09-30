package logic

import (
	"context"
	"strconv"

	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-job/ent"
	"github.com/saas-zero/saas-zero-job/ent/sysjob"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateJobLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateJobLogic {
	return &UpdateJobLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateJob 更新任务（更新后调度器 30s 内自动同步；SyncNow 即时同步）
func (l *UpdateJobLogic) UpdateJob(in *apps.Job) (*apps.EmptyResp, error) {
	if in == nil || in.Id == "" {
		return nil, errInvalidParam("id is required")
	}
	id, err := strconv.ParseInt(in.Id, 10, 64)
	if err != nil {
		return nil, errInvalidParam("invalid id: %s", in.Id)
	}
	existed, err := l.svcCtx.DB.SysJob.Get(l.ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errInvalidParam("job not found")
		}
		return nil, err
	}
	if !existed.DeletedAt.IsZero() {
		return nil, errInvalidParam("job not found")
	}
	if existed.Status == sysjob.StatusSuspended {
		return nil, errInvalidParam("job is suspended, cannot update")
	}
	if in.Handler != "" {
		if !handlerRegistered(l.svcCtx.Scheduler.Registry(), in.Handler) {
			return nil, errInvalidParam("handler %q not registered", in.Handler)
		}
	}
	if in.CronExpression != "" {
		if err := validateCron(in.CronExpression); err != nil {
			return nil, errInvalidParam("invalid cron expression: %v", err)
		}
	}
	if err := validateParams(in.Params); err != nil {
		return nil, err
	}

	upd := l.svcCtx.DB.SysJob.UpdateOneID(id)
	if in.Name != "" {
		upd = upd.SetName(in.Name)
	}
	if in.Group != "" {
		upd = upd.SetGroup(in.Group)
	}
	if in.Handler != "" {
		upd = upd.SetHandler(in.Handler)
	}
	if in.Params != "" {
		upd = upd.SetParams(in.Params)
	}
	if in.CronExpression != "" {
		upd = upd.SetCronExpression(in.CronExpression)
	}
	if in.TimeZone != "" {
		upd = upd.SetTimeZone(in.TimeZone)
	}
	if in.MisfirePolicy != "" {
		upd = upd.SetMisfirePolicy(sysjob.MisfirePolicy(in.MisfirePolicy))
	}
	if in.Status != "" {
		if in.Status != string(sysjob.StatusActive) && in.Status != string(sysjob.StatusInactive) && in.Status != string(sysjob.StatusSuspended) {
			return nil, errInvalidParam("invalid status: %s", in.Status)
		}
		upd = upd.SetStatus(sysjob.Status(in.Status))
	}
	if in.Timeout != 0 {
		upd = upd.SetTimeout(in.Timeout)
	}
	if in.MaxRetry != 0 {
		upd = upd.SetMaxRetry(in.MaxRetry)
	}
	if in.RetryInterval != 0 {
		upd = upd.SetRetryInterval(in.RetryInterval)
	}
	// concurrent 是 bool，proto3 无 presence 语义，无法区分“未传”与 false；
	// 管理端表单总会携带该值（默认 false=不允许并发），因此无条件应用。
	// 直连 gRPC 的调用方需显式传 concurrent。
	upd = upd.SetConcurrent(in.Concurrent).SetRemark(in.Remark)

	ctx := systemCtx(l.ctx)
	ctx = mixins.SetCurrentTenantId(ctx, 0)
	if _, err := upd.Save(ctx); err != nil {
		l.Logger.Errorf("update job failed: id=%d err=%v", id, err)
		return nil, err
	}

	l.svcCtx.Scheduler.SyncNow()
	l.Logger.Infof("job updated: id=%d", id)
	return emptyResp(), nil
}

// emptyResp 统一空响应
func emptyResp() *apps.EmptyResp {
	return &apps.EmptyResp{Code: 0, Msg: "success"}
}
