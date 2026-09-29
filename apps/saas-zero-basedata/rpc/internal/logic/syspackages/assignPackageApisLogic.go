package syspackageslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type AssignPackageApisLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignPackageApisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignPackageApisLogic {
	return &AssignPackageApisLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AssignPackageApis 设置套餐可用的 API 集合（sys_package_apis）。
// 注意：套餐的 API 关联只是权限模板数据，不写入 Casbin 策略。
// Casbin 策略只由"角色分配 API"生成（assignApisLogic），套餐本身不是授权主体。
func (l *AssignPackageApisLogic) AssignPackageApis(in *apps.PackageReq) (*apps.EmptyResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentUserId(l.ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	err := l.svcCtx.DB.SysPackage.UpdateOneID(in.GetId()).
		ClearApis().
		AddAPIIDs(in.GetApiIds()...).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
