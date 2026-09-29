package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type CreatePackageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreatePackageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePackageLogic {
	return &CreatePackageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePackageLogic) CreatePackage(req *types.PackageReq) (*types.BaseResp, error) {
	rpcReq := &apps.PackageReq{
		Name:   proto.String(req.Name),
		Code:   proto.String(req.Code),
		Status: proto.String(req.Status),
	}
	if req.Sort > 0 {
		rpcReq.Sort = proto.Int32(req.Sort)
	}
	if req.Remark != "" {
		rpcReq.Remark = proto.String(req.Remark)
	}
	resp, err := l.svcCtx.SysPackages.CreatePackage(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
