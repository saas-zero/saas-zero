// Code scaffolded by goctl. Safe to edit.

package logic

import (
	"context"
	"fmt"
	"strconv"

	"github.com/saas-zero/saas-zero-auth/api/internal/svc"
	"github.com/saas-zero/saas-zero-auth/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type PasswordResetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasswordResetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PasswordResetLogic {
	return &PasswordResetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// resetPasswordPerm 是重置密码所需的前端按钮权限码（来自菜单 button 节点 path），
// 与系统的按钮级权限模型一致：分配菜单时自动授权同页面的 button 节点。
const resetPasswordPerm = "system:user:resetPassword"

func (l *PasswordResetLogic) PasswordReset(req *types.PasswordResetReq) (resp *types.BaseResp, err error) {
	userId, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return &types.BaseResp{Code: errno.UserIdFormatErr.Code, Msg: errno.UserIdFormatErr.Msg}, nil
	}

	// 1. 统一会话校验：JWT 签名 + 有效期 + Redis token:<jti> + token_version:<userId>
	claims, ok := validateSession(l.svcCtx.Redis, l.svcCtx.Config.JwtSecret, GetToken(l.ctx))
	if ok != nil {
		return &types.BaseResp{Code: ok.Code, Msg: ok.Msg}, nil
	}

	// 2. 授权校验：只有具备"重置密码"权限（按钮权限码）的角色才能调用，
	//    不能只凭"已登录"判断。权限数据来自当前用户自己的菜单树（GetMenuTree），
	//    与前端按钮显隐逻辑一致。
	authCtx := withAuthContext(l.ctx, l.svcCtx.Config.JwtSecret)
	treeResp, terr := l.svcCtx.SysMenus.GetMenuTree(authCtx, &apps.EmptyReq{})
	if terr != nil {
		return nil, terr
	}
	perms := collectButtonPerms(treeResp.GetData())
	if !containsString(perms, resetPasswordPerm) {
		return &types.BaseResp{Code: errno.Forbidden.Code, Msg: "无重置密码权限"}, nil
	}

	// 3. 边界规则：不能自己重置自己（改自己的密码走 /oauth/password/change）。
	if claims.UserId == userId {
		return &types.BaseResp{Code: errno.InvalidParam.Code, Msg: "不能通过重置接口重置本人密码"}, nil
	}

	// 4. RPC 层按当前租户查询目标用户并校验系统管理员边界（禁止跨租户重置）。
	// 密码由 Basedata RPC 统一哈希，Auth 不传递哈希值，避免重复哈希。
	_, err = l.svcCtx.SysUsers.ResetPassword(authCtx, &apps.UserReq{
		Id:       proto.Int64(userId),
		Password: proto.String(req.NewPassword),
	})
	if err != nil {
		return nil, err
	}
	// 5. 重置成功后递增目标用户 token_version，使所有旧 token 立即失效。
	//    仅操作 ID（userId）进入安全审计日志，绝不记录明文密码或 bcrypt 哈希。
	logx.WithContext(l.ctx).Infof("password reset: operator=%d tenant=%d target=%d",
		claims.UserId, claims.TenantId, userId)
	l.svcCtx.Redis.Incr(fmt.Sprintf("token_version:%d", userId))
	return &types.BaseResp{Code: errno.Success.Code, Msg: "密码重置成功"}, nil
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
