package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-job/ent"
	"github.com/saas-zero/saas-zero-job/ent/sysjob"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetJobListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetJobListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetJobListLogic {
	return &GetJobListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetJobList 任务列表（未删除，支持条件过滤 + 分页）
func (l *GetJobListLogic) GetJobList(in *apps.JobListReq) (*apps.JobListResp, error) {
	page, size := normalizePage(in.GetPage(), in.GetPageSize())

	q := l.svcCtx.DB.SysJob.Query().Where(sysjob.DeletedAtIsNil())
	if in.GetName() != "" {
		q = q.Where(sysjob.NameContains(in.GetName()))
	}
	if in.GetGroup() != "" {
		q = q.Where(sysjob.GroupEQ(in.GetGroup()))
	}
	if in.GetHandler() != "" {
		q = q.Where(sysjob.HandlerEQ(in.GetHandler()))
	}
	if in.GetStatus() != "" {
		q = q.Where(sysjob.StatusEQ(sysjob.Status(in.GetStatus())))
	}

	total, err := q.Count(l.ctx)
	if err != nil {
		l.Logger.Errorf("count jobs failed: %v", err)
		return nil, err
	}
	jobs, err := q.
		Order(ent.Desc(sysjob.FieldCreatedAt)).
		Offset((page - 1) * size).
		Limit(size).
		All(l.ctx)
	if err != nil {
		l.Logger.Errorf("list jobs failed: %v", err)
		return nil, err
	}

	list := make([]*apps.Job, 0, len(jobs))
	for _, j := range jobs {
		list = append(list, jobToProto(j))
	}
	return &apps.JobListResp{List: list, Total: int64(total)}, nil
}
