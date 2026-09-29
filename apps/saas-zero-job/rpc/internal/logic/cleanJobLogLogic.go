package logic

import (
	"context"
	"time"

	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-job/ent/sysjob"
	"github.com/saas-zero/saas-zero-job/ent/sysjoblog"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CleanJobLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCleanJobLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CleanJobLogLogic {
	return &CleanJobLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CleanJobLog 按 job_id 清理任务日志（禁止全表清理）
// job_id 必填；keep_days 为 0 时使用默认 30 天。
func (l *CleanJobLogLogic) CleanJobLog(in *apps.CleanJobLogReq) (*apps.EmptyResp, error) {
	if in == nil || in.GetJobId() == "" {
		return nil, errInvalidParam("job_id is required, prohibit cleaning all logs")
	}
	id, err := parseID(in.GetJobId())
	if err != nil {
		return nil, err
	}
	// 校验任务存在
	exists, err := l.svcCtx.DB.SysJob.Query().
		Where(sysjob.IDEQ(id), sysjob.DeletedAtIsNil()).
		Exist(l.ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errInvalidParam("job not found")
	}

	keepDays := int(in.GetKeepDays())
	if keepDays <= 0 {
		keepDays = 30
	}
	cutoff := time.Now().AddDate(0, 0, -keepDays)

	ctx := mixins.SetCurrentTenantId(l.ctx, 0)
	deleted, err := l.svcCtx.DB.SysJobLog.Delete().
		Where(
			sysjoblog.JobIDEQ(id),
			sysjoblog.CreatedAtLT(cutoff),
		).
		Exec(ctx)
	if err != nil {
		l.Logger.Errorf("clean job logs failed: job=%d err=%v", id, err)
		return nil, err
	}
	l.Logger.Infof("job logs cleaned: job=%d deleted=%d keepDays=%d", id, deleted, keepDays)
	return emptyResp(), nil
}
