package systenantslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantByIdLogic {
	return &GetTenantByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTenantByIdLogic) GetTenantById(in *apps.IdReq) (*apps.TenantResp, error) {
	t, err := l.svcCtx.DB.SysTenant.ActiveQuery().
		Where(systenant.IDEQ(in.GetId())).
		WithSysPackage().
		Only(l.ctx)
	if err != nil {
		return nil, err
	}
	return &apps.TenantResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: tenantToResp(t),
	}, nil
}
