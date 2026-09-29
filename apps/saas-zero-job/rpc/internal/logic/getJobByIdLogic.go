package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-job/ent"
	"github.com/saas-zero/saas-zero-job/ent/sysjob"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetJobByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetJobByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetJobByIdLogic {
	return &GetJobByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetJobById 按 ID 获取任务（未删除）
func (l *GetJobByIdLogic) GetJobById(in *apps.IdReq) (*apps.Job, error) {
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
	return jobToProto(j), nil
}
