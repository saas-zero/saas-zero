package systenantslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantByCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantByCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantByCodeLogic {
	return &GetTenantByCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTenantByCodeLogic) GetTenantByCode(in *apps.TenantReq) (*apps.TenantResp, error) {
	t, err := l.svcCtx.DB.SysTenant.ActiveQuery().
		Where(systenant.CodeEQ(in.GetCode())).
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
