package sysroleslogic

import (
	"context"
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type CreateRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRoleLogic {
	return &CreateRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateRoleLogic) CreateRole(in *apps.RoleReq) (*apps.RoleResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)

	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	if err := checkAssignableMenus(l.svcCtx, ctx, in.GetMenuIds()); err != nil {
		return nil, err
	}

	create := l.svcCtx.DB.SysRole.Create().
		SetName(in.GetName()).
		SetCode(in.GetCode()).
		SetStatus(sysrole.Status(in.GetStatus())).
		SetSort(uint32(in.GetSort()))

	if in.GetRemark() != "" {
		create.SetRemark(in.GetRemark())
	}

	if len(in.GetMenuIds()) > 0 {
		create.AddMenuIDs(in.GetMenuIds()...)
	}
	result, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.RoleResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: &apps.Role{
			Id:     proto.Int64(result.ID),
			IdStr:  proto.String(id.ToString(result.ID)),
			Status: proto.String(string(result.Status)),
		},
	}, nil
}
