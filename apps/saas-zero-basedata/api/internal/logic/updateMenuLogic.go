package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type UpdateMenuLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMenuLogic {
	return &UpdateMenuLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMenuLogic) UpdateMenu(req *types.MenuReq) (*types.BaseResp, error) {
	rpcReq := &apps.MenuReq{Id: proto.Int64(parseId(req.Id))}
	if req.MenuType != "" {
		rpcReq.MenuType = proto.String(req.MenuType)
	}
	if req.Name != "" {
		rpcReq.Name = proto.String(req.Name)
	}
	if req.Path != "" {
		rpcReq.Path = proto.String(req.Path)
	}
	if req.Icon != "" {
		rpcReq.Icon = proto.String(req.Icon)
	}
	if req.Component != "" {
		rpcReq.Component = proto.String(req.Component)
	}
	if req.Status != "" {
		rpcReq.Status = proto.String(req.Status)
	}
	if pid := parseId(req.ParentId); pid > 0 {
		rpcReq.ParentId = proto.Int64(pid)
	}
	if req.Sort > 0 {
		rpcReq.Sort = proto.Int32(req.Sort)
	}
	if req.Remark != "" {
		rpcReq.Remark = proto.String(req.Remark)
	}
	resp, err := l.svcCtx.SysMenus.UpdateMenu(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
