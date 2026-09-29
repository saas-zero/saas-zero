package syspackageslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/syspackage"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPackageByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPackageByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPackageByIdLogic {
	return &GetPackageByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPackageByIdLogic) GetPackageById(in *apps.IdReq) (*apps.PackageResp, error) {
	p, err := l.svcCtx.DB.SysPackage.ActiveQuery().
		Where(syspackage.IDEQ(in.GetId())).
		WithMenus().
		WithApis().
		Only(l.ctx)
	if err != nil {
		return nil, err
	}
	return &apps.PackageResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: packageToResp(p),
	}, nil
}
