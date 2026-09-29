package sysuserslogic

import (
	"context"
	"time"

	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/redis"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteUserLogic) DeleteUser(in *apps.IdsReq) (*apps.EmptyResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	// 防止删除当前登录用户自己；所有目标都必须属于当前租户，避免混合 ID 请求部分成功。
	for _, id := range in.GetIds() {
		if id == mixins.GetCurrentUserId(l.ctx) {
			return nil, errno.New(errno.InvalidParam.Code, "不能删除当前登录用户")
		}
	}
	count, err := l.svcCtx.DB.SysUser.TenantQuery(tenantId).
		Where(sysuser.IDIn(in.GetIds()...)).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	if count != len(in.GetIds()) {
		return nil, errno.New(errno.InvalidParam.Code, "存在不存在或不属于当前租户的用户")
	}

	_, err = l.svcCtx.DB.SysUser.Update().
		Where(
			sysuser.IDIn(in.GetIds()...),
			sysuser.TenantIDEQ(tenantId),
			sysuser.DeletedAtIsNil(),
		).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	// 删除用户后递增 token_version，使其所有旧 token 立即失效，被删除用户无法继续访问
	if err := redis.BumpTokenVersions(l.svcCtx.Redis, in.GetIds()...); err != nil {
		return nil, err
	}
	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
