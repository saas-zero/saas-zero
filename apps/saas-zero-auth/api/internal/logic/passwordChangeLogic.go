// Code scaffolded by goctl. Safe to edit.

package logic

import (
	"context"
	"fmt"

	"github.com/saas-zero/saas-zero-auth/api/internal/svc"
	"github.com/saas-zero/saas-zero-auth/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-common/pkg/bcrypt"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type PasswordChangeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasswordChangeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PasswordChangeLogic {
	return &PasswordChangeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasswordChangeLogic) PasswordChange(req *types.PasswordChangeReq) (resp *types.BaseResp, err error) {
	// 统一会话校验：签名 + 有效期 + Redis JTI + tokenVersion
	claims, ok := validateSession(l.svcCtx.Redis, l.svcCtx.Config.JwtSecret, GetToken(l.ctx))
	if ok != nil {
		return &types.BaseResp{Code: ok.Code, Msg: ok.Msg}, nil
	}
	ctx := withAuthContext(l.ctx, l.svcCtx.Config.JwtSecret)
	userResp, err := l.svcCtx.SysUsers.GetUserByUsername(ctx, &apps.UserReq{Username: proto.String(claims.UserName)})
	if err != nil {
		return nil, err
	}
	user := userResp.GetData()
	if user == nil {
		return &types.BaseResp{Code: errno.UserNotFound.Code, Msg: errno.UserNotFound.Msg}, nil
	}
	if !bcrypt.Verify(req.OldPassword, user.GetPassword()) {
		return &types.BaseResp{Code: errno.OldPasswordWrong.Code, Msg: errno.OldPasswordWrong.Msg}, nil
	}
	// Password is hashed exactly once at the Basedata RPC boundary.
	_, err = l.svcCtx.SysUsers.ResetPassword(ctx, &apps.UserReq{
		Id:       proto.Int64(claims.UserId),
		Password: proto.String(req.NewPassword),
	})
	if err != nil {
		return nil, err
	}
	// 修改密码后递增 token_version，使该用户所有旧 token 失效（强制重新登录）
	l.svcCtx.Redis.Incr(fmt.Sprintf("token_version:%d", claims.UserId))
	return &types.BaseResp{Code: errno.Success.Code, Msg: "密码修改成功"}, nil
}
