package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyApisLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyApisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyApisLogic {
	return &GetMyApisLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyApisLogic) GetMyApis() (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysApis.GetMyApis(l.ctx, &apps.EmptyReq{})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{
		Code: int(resp.Code),
		Msg:  resp.Msg,
		Data: &types.PageResp{List: toAPIInfoList(resp.List), Total: resp.Total},
	}, nil
}