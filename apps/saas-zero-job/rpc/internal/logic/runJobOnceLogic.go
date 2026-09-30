package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-job/ent"
	"github.com/saas-zero/saas-zero-job/ent/sysjob"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type RunJobOnceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunJobOnceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunJobOnceLogic {
	return &RunJobOnceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RunJobOnce 立即执行一次（不受 status 限制，但必须存在且未删除；记录 trigger=manual）
//
// 异步语义：handler 在后台 goroutine 执行，本方法只表示“已提交”，
// 返回值不代表执行成功——结果看 sys_job_logs 与 last_status/last_error。
// 前端按钮提示已按此语义（“已触发执行，请查看任务日志”）。
func (l *RunJobOnceLogic) RunJobOnce(in *apps.IdReq) (*apps.EmptyResp, error) {
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

	// 手动执行异步跑，立即返回（成功/失败见日志与 last_*）
	go func() {
		if err := l.svcCtx.Scheduler.RunNow(context.Background(), id); err != nil {
			l.Logger.Errorf("run job once failed: id=%d err=%v", id, err)
		}
	}()
	l.Logger.Infof("job run once triggered: id=%d name=%s", id, j.Name)
	return emptyResp(), nil
}
