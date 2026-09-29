// Code scaffolded by goctl. Safe to edit.

package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-auth/api/internal/svc"
	"github.com/saas-zero/saas-zero-auth/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type OauthUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOauthUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OauthUserInfoLogic {
	return &OauthUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OauthUserInfoLogic) OauthUserInfo() (resp *types.BaseResp, err error) {
	// 统一会话校验：签名 + 有效期 + Redis JTI + tokenVersion
	claims, ok := validateSession(l.svcCtx.Redis, l.svcCtx.Config.JwtSecret, GetToken(l.ctx))
	if ok != nil {
		return &types.BaseResp{Code: ok.Code, Msg: ok.Msg}, nil
	}
	userResp, err := l.svcCtx.SysUsers.GetUserById(withAuthContext(l.ctx, l.svcCtx.Config.JwtSecret), &apps.IdReq{Id: claims.UserId})
	if err != nil {
		return nil, err
	}
	user := userResp.GetData()
	if user == nil {
		return &types.BaseResp{Code: errno.UserNotFound.Code, Msg: errno.UserNotFound.Msg}, nil
	}
	userData := user
	userData.RoleCodes = claims.RoleCodes
	return &types.BaseResp{
		Code: errno.Success.Code, Msg: errno.Success.Msg, Data: userData,
	}, nil
}
