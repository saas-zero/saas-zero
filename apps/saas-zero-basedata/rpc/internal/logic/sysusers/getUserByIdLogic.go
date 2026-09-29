package sysuserslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByIdLogic {
	return &GetUserByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserByIdLogic) GetUserById(in *apps.IdReq) (*apps.UserResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)

	u, err := l.svcCtx.DB.SysUser.TenantQuery(tenantId).
		Where(sysuser.IDEQ(in.GetId())).
		WithRoles().
		WithSysDept().
		Only(l.ctx)
	if err != nil {
		return nil, err
	}
	return &apps.UserResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: userToResp(u),
	}, nil
}
