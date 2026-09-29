package sysdictdataslogic

import (
	"context"
	"time"

	"github.com/saas-zero/saas-zero-basedata/ent/sysdictdata"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/logic/tenantcheck"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDictDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDictDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDictDataLogic {
	return &DeleteDictDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteDictDataLogic) DeleteDictData(in *apps.IdsReq) (*apps.EmptyResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	// 预取目标数据，确认均对当前租户可见；系统默认数据仅系统管理员可删除。
	// 先去重计数，避免重复 ID 绕过“全部目标均有效”的判断。
	uniqueIDs := make(map[int64]struct{}, len(in.GetIds()))
	for _, id := range in.GetIds() {
		uniqueIDs[id] = struct{}{}
	}
	targets, err := l.svcCtx.DB.SysDictData.TenantAwareQuery(tenantId).
		Where(sysdictdata.IDIn(in.GetIds()...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(targets) != len(uniqueIDs) {
		return nil, errno.New(errno.InvalidParam.Code, "存在不属于当前租户的字典数据")
	}
	for _, d := range targets {
		if d.TenantID == 0 {
			if err := tenantcheck.RequireDefaultTenantAdmin(l.svcCtx, ctx); err != nil {
				return nil, err
			}
			break
		}
	}

	_, err = l.svcCtx.DB.SysDictData.Update().
		Where(
			sysdictdata.IDIn(in.GetIds()...),
			sysdictdata.Or(sysdictdata.TenantIDEQ(tenantId), sysdictdata.TenantIDEQ(0)),
			sysdictdata.DeletedAtIsNil(),
		).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
