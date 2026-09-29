package systenantslogic

import (
	"context"
	"time"

	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	idutil "github.com/saas-zero/saas-zero-common/pkg/id"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTenantLogic {
	return &DeleteTenantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteTenantLogic) DeleteTenant(in *apps.IdsReq) (*apps.EmptyResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentUserId(l.ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	_, err := l.svcCtx.DB.SysTenant.Update().
		Where(systenant.IDIn(in.GetIds()...)).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// 清理该租户 dom 下的 Casbin 策略（孤儿数据），避免残留与累积。
	// 雪花 ID 不复用，重建租户不会串策略；这里是纯数据卫生处理。
	if l.svcCtx.Enforcer != nil {
		for _, id := range in.GetIds() {
			if _, err := l.svcCtx.Enforcer.RemoveFilteredPolicy(1, idutil.ToString(id)); err != nil {
				logx.Errorf("deleteTenant: remove casbin policy for tenant %d failed: %v", id, err)
			}
		}
	}

	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
