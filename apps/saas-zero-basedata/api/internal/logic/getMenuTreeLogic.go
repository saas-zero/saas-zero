package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuTreeLogic {
	return &GetMenuTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMenuTreeLogic) GetMenuTree() (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysMenus.GetMenuTree(l.ctx, &apps.EmptyReq{})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: toMenuInfoList(resp.GetData())}, nil
}
