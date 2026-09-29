package sysdictslogic

import (
	"context"
	"time"

	"github.com/saas-zero/saas-zero-basedata/ent/sysdict"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/logic/tenantcheck"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDictLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDictLogic {
	return &DeleteDictLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteDictLogic) DeleteDict(in *apps.IdsReq) (*apps.EmptyResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	// 预取目标字典，确认均为当前租户可见；系统默认字典仅系统管理员可删除。
	// 先去重计数，避免重复 ID 绕过“全部目标均有效”的判断。
	uniqueIDs := make(map[int64]struct{}, len(in.GetIds()))
	for _, id := range in.GetIds() {
		uniqueIDs[id] = struct{}{}
	}
	targets, err := l.svcCtx.DB.SysDict.TenantAwareQuery(tenantId).
		Where(sysdict.IDIn(in.GetIds()...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(targets) != len(uniqueIDs) {
		return nil, errno.New(errno.InvalidParam.Code, "存在不属于当前租户的字典")
	}
	for _, d := range targets {
		if d.TenantID == 0 {
			if err := tenantcheck.RequireDefaultTenantAdmin(l.svcCtx, ctx); err != nil {
				return nil, err
			}
			break
		}
	}

	_, err = l.svcCtx.DB.SysDict.Update().
		Where(
			sysdict.IDIn(in.GetIds()...),
			sysdict.Or(sysdict.TenantIDEQ(tenantId), sysdict.TenantIDEQ(0)),
			sysdict.DeletedAtIsNil(),
		).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
