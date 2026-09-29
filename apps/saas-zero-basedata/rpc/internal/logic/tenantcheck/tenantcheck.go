package tenantcheck

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
)

// IsDefaultTenantAdmin reports whether the current user is the admin of the
// "default" tenant. The default tenant admin is treated as the system
// super-admin: the only role allowed to maintain system-default (tenant_id=0)
// dicts and other global data.
func IsDefaultTenantAdmin(svcCtx *svc.ServiceContext, ctx context.Context) (bool, error) {
	tenantId := mixins.GetCurrentTenantId(ctx)
	if tenantId <= 0 {
		return false, nil
	}
	tenant, err := svcCtx.DB.SysTenant.Query().
		Where(systenant.IDEQ(tenantId), systenant.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	if tenant.Code != "default" {
		return false, nil
	}
	userId := mixins.GetCurrentUserId(ctx)
	if userId <= 0 {
		return false, nil
	}
	hasAdmin, err := svcCtx.DB.SysUser.TenantQuery(tenantId).
		Where(sysuser.IDEQ(userId)).
		QueryRoles().
		Where(sysrole.CodeEQ("admin"), sysrole.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return false, err
	}
	return hasAdmin, nil
}

// RequireDefaultTenantAdmin fails with errno.Forbidden when the caller is not
// the default tenant admin.
func RequireDefaultTenantAdmin(svcCtx *svc.ServiceContext, ctx context.Context) error {
	ok, err := IsDefaultTenantAdmin(svcCtx, ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errno.New(errno.Forbidden.Code, "仅系统管理员可维护系统默认数据")
	}
	return nil
}
