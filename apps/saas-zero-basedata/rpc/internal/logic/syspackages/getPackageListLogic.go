package syspackageslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/syspackage"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/pagination"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPackageListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPackageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPackageListLogic {
	return &GetPackageListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPackageListLogic) GetPackageList(in *apps.PackagePageReq) (*apps.PackageListResp, error) {
	query := l.svcCtx.DB.SysPackage.ActiveQuery()
	if in.GetName() != "" {
		query = query.Where(syspackage.NameContains(in.GetName()))
	}
	if in.GetCode() != "" {
		query = query.Where(syspackage.CodeContains(in.GetCode()))
	}
	if in.GetStatus() != "" {
		query = query.Where(syspackage.StatusEQ(syspackage.Status(in.GetStatus())))
	}

	total, err := query.Count(l.ctx)
	if err != nil {
		return nil, err
	}

	_, size, offset := pagination.Normalize(int(in.GetPage()), int(in.GetSize()))

	packages, err := query.
		Offset(offset).
		Limit(size).
		Order(ent.Asc(syspackage.FieldSort)).
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*apps.Package, len(packages))
	for i, p := range packages {
		list[i] = packageToResp(p)
	}
	return &apps.PackageListResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		List:  list,
		Total: int64(total),
	}, nil
}
