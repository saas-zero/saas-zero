package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleListLogic {
	return &GetRoleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRoleListLogic) GetRoleList(req *types.RolePageReq) (*types.BaseResp, error) {
	rpcReq := &apps.RolePageReq{
		Page: int32(req.Page),
		Size: int32(req.Size),
	}
	if req.Name != "" {
		rpcReq.Name = &req.Name
	}
	if req.Code != "" {
		rpcReq.Code = &req.Code
	}
	if req.Status != "" {
		rpcReq.Status = &req.Status
	}
	resp, err := l.svcCtx.SysRoles.GetRoleList(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{
		Code: int(resp.Code),
		Msg:  resp.Msg,
		Data: &types.PageResp{List: toRoleInfoList(resp.List), Total: resp.Total},
	}, nil
}
