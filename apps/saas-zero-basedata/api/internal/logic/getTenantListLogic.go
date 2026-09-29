package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantListLogic {
	return &GetTenantListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantListLogic) GetTenantList(req *types.TenantPageReq) (*types.BaseResp, error) {
	rpcReq := &apps.TenantPageReq{Page: int32(req.Page), Size: int32(req.Size)}
	if req.Name != "" {
		rpcReq.Name = &req.Name
	}
	if req.Code != "" {
		rpcReq.Code = &req.Code
	}
	if req.Status != "" {
		rpcReq.Status = &req.Status
	}
	resp, err := l.svcCtx.SysTenants.GetTenantList(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{
		Code: int(resp.Code),
		Msg:  resp.Msg,
		Data: &types.PageResp{List: toTenantInfoList(resp.List), Total: resp.Total},
	}, nil
}
