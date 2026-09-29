package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type CreateMenuLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMenuLogic {
	return &CreateMenuLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateMenuLogic) CreateMenu(req *types.MenuReq) (*types.BaseResp, error) {
	menuType := req.MenuType
	if menuType == "" {
		menuType = "directory"
		l.Logger.Infof("menuType is empty, defaulting to directory, req=%+v", req)
	}
	rpcReq := &apps.MenuReq{
		Name:     proto.String(req.Name),
		MenuType: proto.String(menuType),
		Path:     proto.String(req.Path),
		Icon:     proto.String(req.Icon),
		Status:   proto.String(req.Status),
		Sort:     proto.Int32(req.Sort),
		Hidden:   proto.Bool(req.Hidden),
	}
	if pid := parseId(req.ParentId); pid > 0 {
		rpcReq.ParentId = proto.Int64(pid)
	}
	if req.Component != "" {
		rpcReq.Component = proto.String(req.Component)
	}
	if req.Remark != "" {
		rpcReq.Remark = proto.String(req.Remark)
	}
	resp, err := l.svcCtx.SysMenus.CreateMenu(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
