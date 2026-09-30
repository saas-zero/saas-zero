package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-job/ent"
	"github.com/saas-zero/saas-zero-job/ent/sysjob"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type StartJobLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartJobLogic {
	return &StartJobLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// StartJob 启动任务（暂停/停用 → active，调度器开始调度）
func (l *StartJobLogic) StartJob(in *apps.IdReq) (*apps.EmptyResp, error) {
	if in == nil || in.Id == "" {
		return nil, errInvalidParam("id is required")
	}
	id, err := parseID(in.Id)
	if err != nil {
		return nil, err
	}
	j, err := l.svcCtx.DB.SysJob.Query().
		Where(sysjob.IDEQ(id), sysjob.DeletedAtIsNil()).
		Only(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errInvalidParam("job not found")
		}
		return nil, err
	}

	// 校验再次保证可调度的配置完整性
	if !handlerRegistered(l.svcCtx.Scheduler.Registry(), j.Handler) {
		return nil, errInvalidParam("handler %q not registered, cannot start", j.Handler)
	}
	if err := validateCron(j.CronExpression); err != nil {
		return nil, errInvalidParam("invalid cron expression: %v", err)
	}

	ctx := systemCtx(l.ctx)
	// 重置漏跑窗口：暂停期间不应补跑（misfire 判定以 next_run_at 为基准）
	upd := l.svcCtx.DB.SysJob.UpdateOneID(id).SetStatus(sysjob.StatusActive)
	if next := l.svcCtx.Scheduler.NextRunAt(j); !next.IsZero() {
		upd = upd.SetNextRunAt(next)
	}
	if _, err := upd.Save(ctx); err != nil {
		return nil, err
	}

	l.svcCtx.Scheduler.SyncNow()
	l.Logger.Infof("job started: id=%d name=%s", id, j.Name)
	return emptyResp(), nil
}

type PauseJobLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPauseJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PauseJobLogic {
	return &PauseJobLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// PauseJob 暂停任务（active → suspended，调度器移除）
func (l *PauseJobLogic) PauseJob(in *apps.IdReq) (*apps.EmptyResp, error) {
	if in == nil || in.Id == "" {
		return nil, errInvalidParam("id is required")
	}
	id, err := parseID(in.Id)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.DB.SysJob.Query().
		Where(sysjob.IDEQ(id), sysjob.DeletedAtIsNil()).
		Only(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errInvalidParam("job not found")
		}
		return nil, err
	}

	ctx := systemCtx(l.ctx)
	if _, err := l.svcCtx.DB.SysJob.UpdateOneID(id).
		SetStatus(sysjob.StatusSuspended).
		Save(ctx); err != nil {
		return nil, err
	}

	l.svcCtx.Scheduler.SyncNow()
	l.Logger.Infof("job paused: id=%d", id)
	return emptyResp(), nil
}
