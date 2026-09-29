package sysdictdataslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdict"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdictdata"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/logic/tenantcheck"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDictDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDictDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDictDataLogic {
	return &UpdateDictDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDictDataLogic) UpdateDictData(in *apps.DictDataReq) (*apps.DictDataResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	// 目标字典数据必须对当前租户可见，防止跨租户按全局 ID 更新
	target, err := l.svcCtx.DB.SysDictData.TenantAwareQuery(tenantId).
		Where(sysdictdata.IDEQ(in.GetId())).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errno.New(errno.InvalidParam.Code, "字典数据不存在")
		}
		return nil, err
	}
	// 系统默认字典数据（tenant_id=0）仅系统管理员可维护
	if target.TenantID == 0 {
		if err := tenantcheck.RequireDefaultTenantAdmin(l.svcCtx, ctx); err != nil {
			return nil, err
		}
	}
	// 若同时提交 dictId，需校验其仍对当前租户可见
	if in.DictId != nil && in.GetDictId() > 0 {
		dictInScope, derr := l.svcCtx.DB.SysDict.TenantAwareQuery(tenantId).
			Where(sysdict.IDEQ(in.GetDictId())).
			Exist(ctx)
		if derr != nil {
			return nil, derr
		}
		if !dictInScope {
			return nil, errno.New(errno.InvalidParam.Code, "字典不存在")
		}
	}

	update := l.svcCtx.DB.SysDictData.UpdateOne(target)
	if in.Name != nil {
		update.SetName(in.GetName())
	}
	if in.Key != nil {
		update.SetKey(in.GetKey())
	}
	if in.Value != nil {
		update.SetValue(in.GetValue())
	}
	if in.Status != nil {
		update.SetStatus(sysdictdata.Status(in.GetStatus()))
	}
	if in.Remark != nil {
		update.SetRemark(in.GetRemark())
	}

	result, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	d, err := l.svcCtx.DB.SysDictData.TenantAwareQuery(tenantId).Where(sysdictdata.IDEQ(result.ID)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.DictDataResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: dictDataToResp(d),
	}, nil
}
