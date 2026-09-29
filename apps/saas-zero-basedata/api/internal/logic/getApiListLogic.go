package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetApiListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetApiListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApiListLogic {
	return &GetApiListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetApiListLogic) GetApiList(req *types.ApiPageReq) (*types.BaseResp, error) {
	rpcReq := &apps.ApiPageReq{Page: int32(req.Page), Size: int32(req.Size)}
	if req.ApiName != "" {
		rpcReq.ApiName = &req.ApiName
	}
	if req.ApiPath != "" {
		rpcReq.ApiPath = &req.ApiPath
	}
	if req.ApiType != "" {
		rpcReq.ApiType = &req.ApiType
	}
	if req.Status != "" {
		rpcReq.Status = &req.Status
	}
	resp, err := l.svcCtx.SysApis.GetApiList(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{
		Code: int(resp.Code),
		Msg:  resp.Msg,
		Data: &types.PageResp{List: toAPIInfoList(resp.List), Total: resp.Total},
	}, nil
}
