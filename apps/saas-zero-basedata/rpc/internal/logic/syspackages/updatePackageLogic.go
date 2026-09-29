package syspackageslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/syspackage"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePackageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePackageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePackageLogic {
	return &UpdatePackageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePackageLogic) UpdatePackage(in *apps.PackageReq) (*apps.PackageResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentUserId(l.ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	update := l.svcCtx.DB.SysPackage.UpdateOneID(in.GetId())
	if in.Name != nil {
		update.SetName(in.GetName())
	}
	if in.Code != nil {
		update.SetCode(in.GetCode())
	}
	if in.Status != nil {
		update.SetStatus(syspackage.Status(in.GetStatus()))
	}
	if in.Sort != nil {
		update.SetSort(uint32(in.GetSort()))
	}
	if in.Remark != nil {
		update.SetRemark(in.GetRemark())
	}

	result, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	p, err := l.svcCtx.DB.SysPackage.Query().Where(syspackage.IDEQ(result.ID)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.PackageResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: packageToResp(p),
	}, nil
}
