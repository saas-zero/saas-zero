// Code scaffolded by goctl. Safe to edit.

package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/saas-zero/saas-zero-auth/api/internal/svc"
	"github.com/saas-zero/saas-zero-auth/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/jwt"
	"github.com/zeromicro/go-zero/core/logx"
)

type OauthRefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOauthRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OauthRefreshLogic {
	return &OauthRefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OauthRefreshLogic) OauthRefresh(req *types.OauthRefreshReq) (resp *types.BaseResp, err error) {
	token := req.Token
	if token == "" {
		token = GetToken(l.ctx)
	}
	// 统一会话校验：签名 + 有效期 + Redis JTI + tokenVersion。
	// 角色或权限被收回后 tokenVersion 已递增，旧 token 无法通过刷新续期恢复旧权限。
	claims, ok := validateSession(l.svcCtx.Redis, l.svcCtx.Config.JwtSecret, token)
	if ok != nil {
		return &types.BaseResp{Code: ok.Code, Msg: ok.Msg}, nil
	}

	// 重新从 Basedata 查询当前角色码，刷新后的 token 携带最新权限，避免旧 Claims 残缺
	ctx := withAuthContext(WithToken(l.ctx, token), l.svcCtx.Config.JwtSecret)
	rcResp, rerr := l.svcCtx.SysUsers.GetUserRoleCodes(ctx, &apps.IdReq{Id: claims.UserId})
	if rerr != nil || rcResp == nil {
		return &types.BaseResp{Code: errno.AuthServiceUnavailable.Code, Msg: errno.AuthServiceUnavailable.Msg}, nil
	}
	roleCodes := rcResp.GetCodes()

	// 重新读取当前 tokenVersion；若期间被递增，新 token 携带新版本（旧细分版本自然失效）
	curTV, terr := l.svcCtx.Redis.Get("token_version:" + strconv.FormatInt(claims.UserId, 10))
	if terr != nil || curTV == "" {
		return &types.BaseResp{Code: errno.TokenExpired.Code, Msg: errno.TokenExpired.Msg}, nil
	}
	tv, err := strconv.ParseInt(curTV, 10, 64)
	if err != nil {
		return &types.BaseResp{Code: errno.AuthServiceUnavailable.Code, Msg: errno.AuthServiceUnavailable.Msg}, nil
	}

	newClaims := &jwt.Claims{
		UserId:       claims.UserId,
		TenantId:     claims.TenantId,
		UserName:     claims.UserName,
		RoleCodes:    roleCodes,
		TokenVersion: tv,
	}
	newToken, err := jwt.Sign(l.svcCtx.Config.JwtSecret, newClaims, time.Duration(l.svcCtx.Config.JwtExpire)*time.Second)
	if err != nil {
		return nil, err
	}
	// Store new token in Redis, delete old
	l.svcCtx.Redis.Del("token:" + claims.ID)
	if err := l.svcCtx.Redis.Setex("token:"+newClaims.ID, newToken, int(l.svcCtx.Config.JwtExpire)); err != nil {
		return nil, err
	}
	return &types.BaseResp{
		Code: errno.Success.Code,
		Msg:  errno.Success.Msg,
		Data: map[string]string{"token": newToken},
	}, nil
}
