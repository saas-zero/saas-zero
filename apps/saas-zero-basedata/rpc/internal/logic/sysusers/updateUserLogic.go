package sysuserslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdept"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/redis"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserLogic) UpdateUser(in *apps.UserReq) (*apps.UserResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	// 先按「当前租户 + 未删除」定位目标用户，防止跨租户按全局 ID 更新
	target, err := l.svcCtx.DB.SysUser.TenantQuery(tenantId).
		Where(sysuser.IDEQ(in.GetId())).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errno.UserNotFound
		}
		return nil, err
	}

	// 关联对象租户一致性：部门和角色必须属于当前租户
	if err := checkUserRolesInTenant(l.svcCtx, ctx, tenantId, in.GetRoleIds()); err != nil {
		return nil, errno.New(errno.InvalidParam.Code, err.Error())
	}
	if in.DeptId != nil && in.GetDeptId() > 0 {
		deptInTenant, derr := l.svcCtx.DB.SysDept.TenantQuery(tenantId).
			Where(sysdept.IDEQ(in.GetDeptId())).
			Exist(ctx)
		if derr != nil {
			return nil, derr
		}
		if !deptInTenant {
			return nil, errno.New(errno.InvalidParam.Code, "部门不存在或不属于当前租户")
		}
	}

	update := l.svcCtx.DB.SysUser.UpdateOne(target)
	if in.Nickname != nil {
		update.SetNickname(in.GetNickname())
	}
	if in.Mobile != nil {
		update.SetMobile(in.GetMobile())
	}
	if in.Email != nil {
		update.SetEmail(in.GetEmail())
	}
	if in.DeptId != nil {
		update.SetDeptID(in.GetDeptId())
	}
	if in.Status != nil {
		update.SetStatus(sysuser.Status(in.GetStatus()))
	}
	if in.Remark != nil {
		update.SetRemark(in.GetRemark())
	}

	// 记录更新前的状态，用于判断是否发生"禁用"
	prevStatus := target.Status

	result, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	// 禁用用户（active → inactive）时递增 token_version，使其所有旧 token 立即失效，
	// 被禁用的用户无法继续访问系统。递增失败必须报错，否则禁用不生效。
	if in.Status != nil &&
		prevStatus == sysuser.StatusActive &&
		sysuser.Status(in.GetStatus()) == sysuser.StatusInactive {
		if err := redis.BumpTokenVersion(l.svcCtx.Redis, in.GetId()); err != nil {
			return nil, err
		}
	}

	if len(in.GetRoleIds()) > 0 {
		if err := l.svcCtx.DB.SysUser.UpdateOne(target).
			ClearRoles().
			AddRoleIDs(in.GetRoleIds()...).
			Exec(ctx); err != nil {
			return nil, err
		}
		// 角色变更：递增 token_version 使旧 token 失效（重新登录后权限生效）
		if err := redis.BumpTokenVersion(l.svcCtx.Redis, result.ID); err != nil {
			return nil, err
		}
	}

	u, err := l.svcCtx.DB.SysUser.TenantQuery(tenantId).
		Where(sysuser.IDEQ(result.ID)).
		WithRoles().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return &apps.UserResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg, Data: userToResp(u)}, nil
}
