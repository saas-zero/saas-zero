package sysdictslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdict"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/logic/tenantcheck"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDictLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDictLogic {
	return &UpdateDictLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDictLogic) UpdateDict(in *apps.DictReq) (*apps.DictResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	// 目标字典必须可见于当前租户（系统默认 tenant_id=0 或本租户自定义）
	target, err := l.svcCtx.DB.SysDict.TenantAwareQuery(tenantId).
		Where(sysdict.IDEQ(in.GetId())).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errno.New(errno.InvalidParam.Code, "字典不存在")
		}
		return nil, err
	}
	// 系统默认字典（tenant_id=0）仅允许系统管理员维护
	if target.TenantID == 0 {
		if err := tenantcheck.RequireDefaultTenantAdmin(l.svcCtx, ctx); err != nil {
			return nil, err
		}
	}

	update := l.svcCtx.DB.SysDict.UpdateOne(target)
	if in.Name != nil {
		update.SetName(in.GetName())
	}
	if in.Key != nil {
		update.SetKey(in.GetKey())
	}
	if in.Status != nil {
		update.SetStatus(sysdict.Status(in.GetStatus()))
	}
	if in.Remark != nil {
		update.SetRemark(in.GetRemark())
	}

	result, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	d, err := l.svcCtx.DB.SysDict.TenantAwareQuery(tenantId).Where(sysdict.IDEQ(result.ID)).WithSysDictDatas().Only(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.DictResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: dictToResp(d),
	}, nil
}
