package syspackageslogic

import (
	"context"
	"time"

	"github.com/saas-zero/saas-zero-basedata/ent/syspackage"
	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePackageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeletePackageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePackageLogic {
	return &DeletePackageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeletePackageLogic) DeletePackage(in *apps.IdsReq) (*apps.EmptyResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentUserId(l.ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	for _, id := range in.GetIds() {
		tenantCount, err := l.svcCtx.DB.SysTenant.Query().
			Where(systenant.PackageIDEQ(id), systenant.DeletedAtIsNil()).
			Count(ctx)
		if err != nil {
			return nil, err
		}
		if tenantCount > 0 {
			return &apps.EmptyResp{Code: int32(errno.PackageInUse.Code), Msg: errno.PackageInUse.Msg}, nil
		}
	}

	_, err := l.svcCtx.DB.SysPackage.Update().
		Where(syspackage.IDIn(in.GetIds()...)).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
