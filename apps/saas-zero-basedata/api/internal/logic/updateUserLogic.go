package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UserReq) (*types.BaseResp, error) {
	rpcReq := &apps.UserReq{Id: proto.Int64(parseId(req.Id))}
	if req.Nickname != "" {
		rpcReq.Nickname = proto.String(req.Nickname)
	}
	if req.Mobile != "" {
		rpcReq.Mobile = proto.String(req.Mobile)
	}
	if req.Email != "" {
		rpcReq.Email = proto.String(req.Email)
	}
	if deptId := parseId(req.DeptId); deptId > 0 {
		rpcReq.DeptId = proto.Int64(deptId)
	}
	if req.Status != "" {
		rpcReq.Status = proto.String(req.Status)
	}
	if req.Remark != "" {
		rpcReq.Remark = proto.String(req.Remark)
	}
	if len(req.RoleIds) > 0 {
		rpcReq.RoleIds = parseIds(req.RoleIds)
	}
	resp, err := l.svcCtx.SysUsers.UpdateUser(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
