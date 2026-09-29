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

type GetUserRoleCodesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserRoleCodesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRoleCodesLogic {
	return &GetUserRoleCodesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserRoleCodesLogic) GetUserRoleCodes(in *apps.IdReq) (*apps.RoleCodesResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	u, err := l.svcCtx.DB.SysUser.TenantQuery(tenantId).
		Where(sysuser.IDEQ(in.GetId())).
		WithRoles().
		Only(l.ctx)
	if err != nil {
		return nil, err
	}
	codes := make([]string, len(u.Edges.Roles))
	for i, r := range u.Edges.Roles {
		codes[i] = r.Code
	}
	return &apps.RoleCodesResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		Codes: codes,
	}, nil
}
