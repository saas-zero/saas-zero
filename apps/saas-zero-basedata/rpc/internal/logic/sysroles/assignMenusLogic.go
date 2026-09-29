package sysroleslogic

import (
	"context"
	"fmt"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type AssignMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignMenusLogic {
	return &AssignMenusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AssignMenusLogic) AssignMenus(in *apps.RoleReq) (*apps.EmptyResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	// 系统内置角色不可改菜单权限；目标角色必须先按「当前租户 + 未删除」定位，防止跨租户操作
	role, err := l.svcCtx.DB.SysRole.TenantQuery(tenantId).
		Where(sysrole.IDEQ(in.GetId())).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errno.New(errno.InvalidParam.Code, "角色不存在或不属于当前租户")
		}
		return nil, err
	}
	if role.IsSystem {
		return nil, errno.New(errno.InvalidParam.Code, fmt.Sprintf("系统内置角色「%s」不可修改", role.Name))
	}

	// 继承式授权：只能把当前用户自己拥有的菜单授给别人
	if err := checkAssignableMenus(l.svcCtx, l.ctx, in.GetMenuIds()); err != nil {
		return nil, err
	}

	err = l.svcCtx.DB.SysRole.UpdateOne(role).
		ClearMenus().
		AddMenuIDs(in.GetMenuIds()...).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
