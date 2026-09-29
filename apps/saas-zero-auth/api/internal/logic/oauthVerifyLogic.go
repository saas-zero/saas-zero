// Code scaffolded by goctl. Safe to edit.

package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-auth/api/internal/svc"
	"github.com/saas-zero/saas-zero-auth/api/internal/types"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type OauthVerifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOauthVerifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OauthVerifyLogic {
	return &OauthVerifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OauthVerifyLogic) OauthVerify() (resp *types.BaseResp, err error) {
	// 统一会话校验：签名 + 有效期 + Redis JTI + tokenVersion
	claims, ok := validateSession(l.svcCtx.Redis, l.svcCtx.Config.JwtSecret, GetToken(l.ctx))
	if ok != nil {
		return &types.BaseResp{Code: ok.Code, Msg: ok.Msg}, nil
	}
	return &types.BaseResp{
		Code: errno.Success.Code,
		Msg:  errno.Success.Msg,
		Data: map[string]interface{}{
			"userId":    claims.UserId,
			"tenantId":  claims.TenantId,
			"userName":  claims.UserName,
			"roleCodes": claims.RoleCodes,
		},
	}, nil
}
