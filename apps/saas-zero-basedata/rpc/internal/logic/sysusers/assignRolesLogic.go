package sysuserslogic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/redis"
)

type AssignRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRolesLogic {
	return &AssignRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AssignRolesLogic) AssignRoles(in *apps.UserReq) (*apps.EmptyResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	// 目标用户必须属于当前租户
	target, err := l.svcCtx.DB.SysUser.TenantQuery(tenantId).
		Where(sysuser.IDEQ(in.GetId())).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errno.UserNotFound
		}
		return nil, err
	}

	// 分配的角色必须属于当前租户，禁止跨租户关联
	if err := checkUserRolesInTenant(l.svcCtx, ctx, tenantId, in.GetRoleIds()); err != nil {
		return nil, errno.New(errno.InvalidParam.Code, err.Error())
	}

	err = l.svcCtx.DB.SysUser.UpdateOne(target).
		ClearRoles().
		AddRoleIDs(in.GetRoleIds()...).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	// 角色变更后旧 token 里的 roleCodes 已过期，必须递增 token_version 强制重新登录。
	if err := redis.BumpTokenVersion(l.svcCtx.Redis, in.GetId()); err != nil {
		return nil, err
	}
	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
