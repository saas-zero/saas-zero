package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type UpdateDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDeptLogic {
	return &UpdateDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDeptLogic) UpdateDept(req *types.DeptReq) (*types.BaseResp, error) {
	rpcReq := &apps.DeptReq{Id: proto.Int64(parseId(req.Id))}
	if req.Name != "" {
		rpcReq.Name = proto.String(req.Name)
	}
	if req.Status != "" {
		rpcReq.Status = proto.String(req.Status)
	}
	if pid := parseId(req.ParentId); pid > 0 {
		rpcReq.ParentId = proto.Int64(pid)
	}
	if lid := parseId(req.LeaderId); lid > 0 {
		rpcReq.LeaderId = proto.Int64(lid)
	}
	if req.Mobile != "" {
		rpcReq.Mobile = proto.String(req.Mobile)
	}
	if req.Email != "" {
		rpcReq.Email = proto.String(req.Email)
	}
	if req.Sort > 0 {
		rpcReq.Sort = proto.Int32(req.Sort)
	}
	resp, err := l.svcCtx.SysDepts.UpdateDept(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
