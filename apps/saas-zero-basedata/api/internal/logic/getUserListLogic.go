package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserListLogic {
	return &GetUserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserListLogic) GetUserList(req *types.UserPageReq) (*types.BaseResp, error) {
	rpcReq := &apps.UserPageReq{
		Page: int32(req.Page),
		Size: int32(req.Size),
	}
	if req.Username != "" {
		rpcReq.Username = &req.Username
	}
	if req.Nickname != "" {
		rpcReq.Nickname = &req.Nickname
	}
	if req.Mobile != "" {
		rpcReq.Mobile = &req.Mobile
	}
	if req.Status != "" {
		rpcReq.Status = &req.Status
	}
	if req.DeptId > 0 {
		rpcReq.DeptId = &req.DeptId
	}
	resp, err := l.svcCtx.SysUsers.GetUserList(l.ctx, rpcReq)
	if err != nil {
		logx.Errorf("GetUserList gRPC error: %v", err)
		return nil, err
	}
	return &types.BaseResp{
		Code: int(resp.Code),
		Msg:  resp.Msg,
		Data: &types.PageResp{List: toSysUserList(resp.List), Total: resp.Total},
	}, nil
}
