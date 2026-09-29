package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePackageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeletePackageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePackageLogic {
	return &DeletePackageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePackageLogic) DeletePackage(req *types.IdsReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysPackages.DeletePackage(l.ctx, &apps.IdsReq{Ids: parseIds(req.Ids)})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}
