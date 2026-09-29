package sysroleslogic

import (
	"context"

	"fmt"
	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/id"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateRoleLogic) UpdateRole(in *apps.RoleReq) (*apps.RoleResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	// 记录变更前信息：旧 code（用于同步 Casbin）、旧状态（用于禁用踢出）。
	// 先按「当前租户 + 未删除」定位角色，防止跨租户按全局 ID 更新。
	oldRole, err := l.svcCtx.DB.SysRole.TenantQuery(tenantId).
		Where(sysrole.IDEQ(in.GetId())).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errno.New(errno.InvalidParam.Code, "角色不存在")
		}
		return nil, err
	}
	oldCode := oldRole.Code
	oldStatus := oldRole.Status

	// 菜单关联必须在角色字段落库前完成授权校验，避免非法请求留下部分更新。
	if len(in.GetMenuIds()) > 0 {
		if err := checkAssignableMenus(l.svcCtx, ctx, in.GetMenuIds()); err != nil {
			return nil, err
		}
	}

	// 系统内置角色（is_system=true）不可修改
	if oldRole.IsSystem {
		return nil, errno.New(errno.InvalidParam.Code, fmt.Sprintf("系统内置角色「%s」不可修改", oldRole.Name))
	}

	update := l.svcCtx.DB.SysRole.UpdateOne(oldRole)
	if in.Name != nil {
		update.SetName(in.GetName())
	}
	if in.Code != nil {
		update.SetCode(in.GetCode())
	}
	if in.Status != nil {
		update.SetStatus(sysrole.Status(in.GetStatus()))
	}
	if in.Sort != nil {
		update.SetSort(uint32(in.GetSort()))
	}
	if in.Remark != nil {
		update.SetRemark(in.GetRemark())
	}

	result, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	// 修改角色 code：同步 Casbin 策略 sub（删旧 code 策略，重建为新 code）
	if in.Code != nil && in.GetCode() != oldCode {
		dom := id.ToString(mixins.GetCurrentTenantId(ctx))
		if l.svcCtx.Enforcer != nil {
			oldPolicies, _ := l.svcCtx.Enforcer.GetFilteredPolicy(0, oldCode, dom)
			l.svcCtx.Enforcer.RemoveFilteredPolicy(0, oldCode, dom)
			for _, p := range oldPolicies {
				if len(p) >= 5 {
					l.svcCtx.Enforcer.AddPolicy(in.GetCode(), dom, p[2], p[3], p[4])
				}
			}
		}
	}

	// 禁用角色（active → inactive）：踢掉拥有该角色的所有用户
	if in.Status != nil && sysrole.Status(in.GetStatus()) == sysrole.StatusInactive && oldStatus == sysrole.StatusActive {
		users, err := l.svcCtx.DB.SysUser.Query().
			Where(sysuser.HasRolesWith(sysrole.IDEQ(result.ID), sysrole.DeletedAtIsNil())).
			All(ctx)
		if err == nil {
			// 角色被禁用后，其用户的旧 token 仍携带该角色，必须强制失效
			if err := svc.BumpUsersTokenVersion(l.svcCtx.Redis, users); err != nil {
				return nil, err
			}
		}
	}

	if len(in.GetMenuIds()) > 0 {
		if err := l.svcCtx.DB.SysRole.UpdateOne(result).
			ClearMenus().
			AddMenuIDs(in.GetMenuIds()...).
			Exec(ctx); err != nil {
			return nil, err
		}
	}
	r, err := l.svcCtx.DB.SysRole.TenantQuery(tenantId).Where(sysrole.IDEQ(result.ID)).WithMenus().Only(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.RoleResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: roleToResp(r),
	}, nil
}
