package sysdeptslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/sysdept"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeptByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDeptByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeptByIdLogic {
	return &GetDeptByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDeptByIdLogic) GetDeptById(in *apps.IdReq) (*apps.DeptResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)

	d, err := l.svcCtx.DB.SysDept.TenantQuery(tenantId).
		Where(sysdept.IDEQ(in.GetId())).
		WithLeader().
		Only(l.ctx)
	if err != nil {
		return nil, err
	}
	return &apps.DeptResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: deptToResp(d),
	}, nil
}
