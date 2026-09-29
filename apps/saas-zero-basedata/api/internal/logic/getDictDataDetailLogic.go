package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictDataDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDictDataDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictDataDetailLogic {
	return &GetDictDataDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDictDataDetailLogic) GetDictDataDetail(req *types.IdReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysDictDatas.GetDictDataById(l.ctx, &apps.IdReq{Id: parseId(req.Id)})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: toDictDataInfo(resp.GetData())}, nil
}
