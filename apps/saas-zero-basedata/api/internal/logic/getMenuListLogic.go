package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMenuListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuListLogic {
	return &GetMenuListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMenuListLogic) GetMenuList(req *types.MenuPageReq) (*types.BaseResp, error) {
	rpcReq := &apps.MenuPageReq{Page: int32(req.Page), Size: int32(req.Size)}
	if req.Name != "" {
		rpcReq.Name = &req.Name
	}
	if req.Status != "" {
		rpcReq.Status = &req.Status
	}
	resp, err := l.svcCtx.SysMenus.GetMenuList(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{
		Code: int(resp.Code),
		Msg:  resp.Msg,
		Data: &types.PageResp{List: toMenuInfoList(resp.List), Total: resp.Total},
	}, nil
}
