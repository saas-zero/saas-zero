package systenantslogic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
)

type ChangeTenantStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangeTenantStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeTenantStatusLogic {
	return &ChangeTenantStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChangeTenantStatusLogic) ChangeTenantStatus(in *apps.TenantReq) (*apps.EmptyResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentUserId(l.ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	err := l.svcCtx.DB.SysTenant.UpdateOneID(in.GetId()).
		SetStatus(systenant.Status(in.GetStatus())).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
