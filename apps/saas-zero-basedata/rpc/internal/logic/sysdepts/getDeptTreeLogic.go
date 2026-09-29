package sysdeptslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdept"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeptTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDeptTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeptTreeLogic {
	return &GetDeptTreeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDeptTreeLogic) GetDeptTree(_ *apps.EmptyReq) (*apps.DeptTreeResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)

	allDepts, err := l.svcCtx.DB.SysDept.TenantQuery(tenantId).
		Order(ent.Asc(sysdept.FieldSort)).
		WithLeader().
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	tree := buildDeptTree(allDepts, 0)
	return &apps.DeptTreeResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: tree,
	}, nil
}
