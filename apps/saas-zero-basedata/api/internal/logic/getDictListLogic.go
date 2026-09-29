package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDictListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictListLogic {
	return &GetDictListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDictListLogic) GetDictList(req *types.DictPageReq) (*types.BaseResp, error) {
	rpcReq := &apps.DictPageReq{Page: int32(req.Page), Size: int32(req.Size)}
	if req.Name != "" {
		rpcReq.Name = &req.Name
	}
	if req.Key != "" {
		rpcReq.Key = &req.Key
	}
	if req.Status != "" {
		rpcReq.Status = &req.Status
	}
	resp, err := l.svcCtx.SysDicts.GetDictList(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{
		Code: int(resp.Code),
		Msg:  resp.Msg,
		Data: &types.PageResp{List: toDictInfoList(resp.List), Total: resp.Total},
	}, nil
}
