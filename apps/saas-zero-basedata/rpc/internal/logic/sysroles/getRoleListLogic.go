package sysroleslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/pagination"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleListLogic {
	return &GetRoleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRoleListLogic) GetRoleList(in *apps.RolePageReq) (*apps.RoleListResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)

	query := l.svcCtx.DB.SysRole.TenantQuery(tenantId)
	if in.GetName() != "" {
		query = query.Where(sysrole.NameContains(in.GetName()))
	}
	if in.GetCode() != "" {
		query = query.Where(sysrole.CodeContains(in.GetCode()))
	}
	if in.GetStatus() != "" {
		query = query.Where(sysrole.StatusEQ(sysrole.Status(in.GetStatus())))
	}

	total, err := query.Count(l.ctx)
	if err != nil {
		return nil, err
	}

	_, size, offset := pagination.Normalize(int(in.GetPage()), int(in.GetSize()))

	roles, err := query.
		Offset(offset).
		Limit(size).
		Order(ent.Asc(sysrole.FieldSort)).
		WithMenus().
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*apps.Role, len(roles))
	for i, r := range roles {
		resp := roleToResp(r)
		resp.ApiIds = roleApiIds(l.svcCtx.Enforcer, r.Code, tenantId)
		list[i] = resp
	}
	return &apps.RoleListResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		List:  list,
		Total: int64(total),
	}, nil
}
