package logic

import (
	"context"
	"time"

	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-job/ent"
	"github.com/saas-zero/saas-zero-job/ent/sysjob"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteJobLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteJobLogic {
	return &DeleteJobLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteJob 软删任务（仅允许删除 suspended/inactive 任务；做到一半提醒，不强制）
func (l *DeleteJobLogic) DeleteJob(in *apps.IdsReq) (*apps.EmptyResp, error) {
	if in == nil || len(in.Ids) == 0 {
		return nil, errInvalidParam("ids is required")
	}
	// 先校验所有任务都存在且未被删除
	for _, sid := range in.Ids {
		id, err := parseID(sid)
		if err != nil {
			return nil, err
		}
		j, err := l.svcCtx.DB.SysJob.Query().
			Where(
				sysjob.IDEQ(id),
				sysjob.DeletedAtIsNil(),
			).Only(l.ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, errInvalidParam("job %s not found", sid)
			}
			return nil, err
		}
		if j.Status == sysjob.StatusActive {
			return nil, errInvalidParam("job %s is active, pause it first", sid)
		}
	}

	ctx := systemCtx(l.ctx)
	ctx = mixins.SetCurrentTenantId(ctx, 0)
	now := time.Now()
	for _, sid := range in.Ids {
		id, _ := parseID(sid)
		if _, err := l.svcCtx.DB.SysJob.UpdateOneID(id).
			SetDeletedAt(now).
			SetStatus(sysjob.StatusInactive).
			Save(ctx); err != nil {
			l.Logger.Errorf("delete job failed: id=%d err=%v", id, err)
			return nil, err
		}
	}

	l.svcCtx.Scheduler.SyncNow()
	l.Logger.Infof("jobs deleted: %v", in.Ids)
	return emptyResp(), nil
}
