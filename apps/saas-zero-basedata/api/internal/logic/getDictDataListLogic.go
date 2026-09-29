package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictDataListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDictDataListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictDataListLogic {
	return &GetDictDataListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDictDataListLogic) GetDictDataList(req *types.DictDataPageReq) (*types.BaseResp, error) {
	rpcReq := &apps.DictDataPageReq{Page: int32(req.Page), Size: int32(req.Size)}
	if req.DictId > 0 {
		rpcReq.DictId = &req.DictId
	}
	if req.Key != "" {
		rpcReq.Key = &req.Key
	}
	if req.Value != "" {
		rpcReq.Value = &req.Value
	}
	if req.Status != "" {
		rpcReq.Status = &req.Status
	}
	resp, err := l.svcCtx.SysDictDatas.GetDictDataList(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{
		Code: int(resp.Code),
		Msg:  resp.Msg,
		Data: &types.PageResp{List: toDictDataInfoList(resp.List), Total: resp.Total},
	}, nil
}
