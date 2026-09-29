// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"fmt"

	"github.com/saas-zero/saas-zero-auth/api/internal/svc"
	"github.com/saas-zero/saas-zero-auth/api/internal/types"
	"github.com/saas-zero/saas-zero-common/pkg/captcha"
	"github.com/saas-zero/saas-zero-common/pkg/errno"

	"github.com/zeromicro/go-zero/core/logx"
)

const captchaTTL = 300

type OauthCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOauthCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OauthCodeLogic {
	return &OauthCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OauthCodeLogic) OauthCode() (resp *types.BaseResp, err error) {
	r, err := captcha.Generate()
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.Redis.Setex(fmt.Sprintf("captcha:%s", r.Id), r.Code, captchaTTL); err != nil {
		return nil, err
	}
	return &types.BaseResp{
		Code: errno.Success.Code,
		Msg:  errno.Success.Msg,
		Data: map[string]string{
			"captchaId":  r.Id,
			"captchaImg": r.B64s,
		},
	}, nil
}
