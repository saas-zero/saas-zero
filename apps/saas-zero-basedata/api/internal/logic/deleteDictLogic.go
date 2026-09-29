package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDictLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteDictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDictLogic {
	return &DeleteDictLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteDictLogic) DeleteDict(req *types.IdsReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysDicts.DeleteDict(l.ctx, &apps.IdsReq{Ids: parseIds(req.Ids)})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}
