package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-job/ent"
	"github.com/saas-zero/saas-zero-job/ent/sysjoblog"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetJobLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetJobLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetJobLogListLogic {
	return &GetJobLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetJobLogList 任务日志列表（必须按 jobId 或 jobName 过滤，支持状态/时间范围 + 分页）
func (l *GetJobLogListLogic) GetJobLogList(in *apps.JobLogListReq) (*apps.JobLogListResp, error) {
	page, size := normalizePage(in.GetPage(), in.GetPageSize())

	q := l.svcCtx.DB.SysJobLog.Query()
	if in.GetJobId() != "" {
		id, err := parseID(in.GetJobId())
		if err != nil {
			return nil, err
		}
		q = q.Where(sysjoblog.JobIDEQ(id))
	}
	if in.GetJobName() != "" {
		q = q.Where(sysjoblog.JobNameContains(in.GetJobName()))
	}
	if in.GetStatus() != "" {
		q = q.Where(sysjoblog.StatusEQ(sysjoblog.Status(in.GetStatus())))
	}
	if in.GetBeginTime() > 0 {
		q = q.Where(sysjoblog.CreatedAtGTE(unixTime(in.GetBeginTime())))
	}
	if in.GetEndTime() > 0 {
		q = q.Where(sysjoblog.CreatedAtLTE(unixTime(in.GetEndTime())))
	}

	total, err := q.Count(l.ctx)
	if err != nil {
		l.Logger.Errorf("count job logs failed: %v", err)
		return nil, err
	}
	logs, err := q.
		Order(ent.Desc(sysjoblog.FieldCreatedAt)).
		Offset((page - 1) * size).
		Limit(size).
		All(l.ctx)
	if err != nil {
		l.Logger.Errorf("list job logs failed: %v", err)
		return nil, err
	}

	list := make([]*apps.JobLog, 0, len(logs))
	for _, lg := range logs {
		list = append(list, logToProto(lg))
	}
	return &apps.JobLogListResp{List: list, Total: int64(total)}, nil
}
