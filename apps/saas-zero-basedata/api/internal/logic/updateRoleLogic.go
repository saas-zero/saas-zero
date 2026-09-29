package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type UpdateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRoleLogic) UpdateRole(req *types.RoleReq) (*types.BaseResp, error) {
	rpcReq := &apps.RoleReq{Id: proto.Int64(parseId(req.Id))}
	if req.Name != "" {
		rpcReq.Name = proto.String(req.Name)
	}
	if req.Code != "" {
		rpcReq.Code = proto.String(req.Code)
	}
	if req.Status != "" {
		rpcReq.Status = proto.String(req.Status)
	}
	if req.Sort > 0 {
		rpcReq.Sort = proto.Int32(req.Sort)
	}
	if req.Remark != "" {
		rpcReq.Remark = proto.String(req.Remark)
	}
	if len(req.MenuIds) > 0 {
		rpcReq.MenuIds = parseIds(req.MenuIds)
	}
	if len(req.ApiIds) > 0 {
		rpcReq.ApiIds = parseIds(req.ApiIds)
	}
	resp, err := l.svcCtx.SysRoles.UpdateRole(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
