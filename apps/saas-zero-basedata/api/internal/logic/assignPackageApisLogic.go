package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type AssignPackageApisLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignPackageApisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignPackageApisLogic {
	return &AssignPackageApisLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssignPackageApisLogic) AssignPackageApis(req *types.PackageReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysPackages.AssignPackageApis(l.ctx, &apps.PackageReq{
		Id:     proto.Int64(parseId(req.Id)),
		ApiIds: parseIds(req.ApiIds),
	})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}
