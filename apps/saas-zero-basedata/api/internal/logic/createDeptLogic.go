package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type CreateDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDeptLogic {
	return &CreateDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDeptLogic) CreateDept(req *types.DeptReq) (*types.BaseResp, error) {
	rpcReq := &apps.DeptReq{
		Name:   proto.String(req.Name),
		Status: proto.String(req.Status),
		Sort:   proto.Int32(req.Sort),
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
	resp, err := l.svcCtx.SysDepts.CreateDept(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
