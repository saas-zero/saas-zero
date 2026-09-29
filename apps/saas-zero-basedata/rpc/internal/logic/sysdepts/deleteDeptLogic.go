package sysdeptslogic

import (
	"context"
	"time"

	"github.com/saas-zero/saas-zero-basedata/ent/sysdept"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDeptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDeptLogic {
	return &DeleteDeptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteDeptLogic) DeleteDept(in *apps.IdsReq) (*apps.EmptyResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	count, err := l.svcCtx.DB.SysDept.TenantQuery(tenantId).
		Where(sysdept.IDIn(in.GetIds()...)).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	if count != len(in.GetIds()) {
		return nil, errno.New(errno.InvalidParam.Code, "存在不存在或不属于当前租户的部门")
	}

	for _, id := range in.GetIds() {
		childCount, err := l.svcCtx.DB.SysDept.Query().
			Where(sysdept.ParentIDEQ(id), sysdept.DeletedAtIsNil(), sysdept.TenantIDEQ(tenantId)).
			Count(ctx)
		if err != nil {
			return nil, err
		}
		if childCount > 0 {
			return &apps.EmptyResp{Code: int32(errno.HasSubDept.Code), Msg: errno.HasSubDept.Msg}, nil
		}
	}

	_, err = l.svcCtx.DB.SysDept.Update().
		Where(
			sysdept.IDIn(in.GetIds()...),
			sysdept.TenantIDEQ(tenantId),
			sysdept.DeletedAtIsNil(),
		).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
