package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPackageDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPackageDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPackageDetailLogic {
	return &GetPackageDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPackageDetailLogic) GetPackageDetail(req *types.IdReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysPackages.GetPackageById(l.ctx, &apps.IdReq{Id: parseId(req.Id)})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: toPackageInfo(resp.GetData())}, nil
}
