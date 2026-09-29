package syspackageslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type AssignPackageMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignPackageMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignPackageMenusLogic {
	return &AssignPackageMenusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AssignPackageMenus 设置套餐可用的菜单集合（sys_package_menus）。
// 套餐是租户的权限模板，运行时权限仍由角色决定，这里只维护模板数据。
func (l *AssignPackageMenusLogic) AssignPackageMenus(in *apps.PackageReq) (*apps.EmptyResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentUserId(l.ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	err := l.svcCtx.DB.SysPackage.UpdateOneID(in.GetId()).
		ClearMenus().
		AddMenuIDs(in.GetMenuIds()...).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
