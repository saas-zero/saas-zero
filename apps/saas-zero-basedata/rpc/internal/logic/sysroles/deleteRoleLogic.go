package sysroleslogic

import (
	"context"
	"fmt"
	"time"

	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/id"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteRoleLogic) DeleteRole(in *apps.IdsReq) (*apps.EmptyResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentUserId(l.ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	tenantId := mixins.GetCurrentTenantId(ctx)

	// 记录待删角色的 code（用于 Casbin 清理）与 id（用于 Token 失效）
	roles, err := l.svcCtx.DB.SysRole.Query().
		Where(
			sysrole.IDIn(in.GetIds()...),
			sysrole.TenantIDEQ(tenantId),
			sysrole.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	if len(roles) != len(in.GetIds()) {
		return nil, errno.New(errno.InvalidParam.Code, "存在不存在或不属于当前租户的角色")
	}

	// 系统内置角色（is_system=true）不可删除
	for _, r := range roles {
		if r.IsSystem {
			return nil, errno.New(errno.InvalidParam.Code, fmt.Sprintf("系统内置角色「%s」不可删除", r.Name))
		}
	}

	_, err = l.svcCtx.DB.SysRole.Update().
		Where(
			sysrole.IDIn(in.GetIds()...),
			sysrole.TenantIDEQ(tenantId),
		).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// 清理 Casbin 策略（本租户 dom 下该角色 code 的所有 API 权限）
	for _, r := range roles {
		if l.svcCtx.Enforcer != nil {
			l.svcCtx.Enforcer.RemoveFilteredPolicy(0, r.Code, id.ToString(tenantId))
		}
	}

	// 该角色下所有用户 Token 失效，踢掉旧会话
	for _, r := range roles {
		users, err := l.svcCtx.DB.SysUser.Query().
			Where(sysuser.TenantIDEQ(tenantId), sysuser.DeletedAtIsNil()).
			Where(sysuser.HasRolesWith(sysrole.IDEQ(r.ID), sysrole.DeletedAtIsNil())).
			All(ctx)
		if err != nil {
			continue
		}
		if err := svc.BumpUsersTokenVersion(l.svcCtx.Redis, users); err != nil {
			return nil, err
		}
	}

	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
