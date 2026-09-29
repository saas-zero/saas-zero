package ent

import (
	"github.com/saas-zero/saas-zero-basedata/ent/sysapi"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdept"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdict"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdictdata"
	"github.com/saas-zero/saas-zero-basedata/ent/sysmenu"
	"github.com/saas-zero/saas-zero-basedata/ent/syspackage"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
)

// ─────────────────────────────────────────────
// ActiveQuery - 过滤已删除记录 (WHERE deleted_at IS NULL)
// ─────────────────────────────────────────────

func (c *SysUserClient) ActiveQuery() *SysUserQuery {
	return c.Query().Where(sysuser.DeletedAtIsNil())
}

func (c *SysTenantClient) ActiveQuery() *SysTenantQuery {
	return c.Query().Where(systenant.DeletedAtIsNil())
}

func (c *SysDeptClient) ActiveQuery() *SysDeptQuery {
	return c.Query().Where(sysdept.DeletedAtIsNil())
}

func (c *SysRoleClient) ActiveQuery() *SysRoleQuery {
	return c.Query().Where(sysrole.DeletedAtIsNil())
}

func (c *SysMenuClient) ActiveQuery() *SysMenuQuery {
	return c.Query().Where(sysmenu.DeletedAtIsNil())
}

func (c *SysDictClient) ActiveQuery() *SysDictQuery {
	return c.Query().Where(sysdict.DeletedAtIsNil())
}

func (c *SysDictDataClient) ActiveQuery() *SysDictDataQuery {
	return c.Query().Where(sysdictdata.DeletedAtIsNil())
}

func (c *SysApiClient) ActiveQuery() *SysApiQuery {
	return c.Query().Where(sysapi.DeletedAtIsNil())
}

func (c *SysPackageClient) ActiveQuery() *SysPackageQuery {
	return c.Query().Where(syspackage.DeletedAtIsNil())
}

// ─────────────────────────────────────────────
// TenantQuery - 当前租户的未删除数据
// 适用于 tenant_id 必填的表（sysuser, sysdept, sysrole）
// ─────────────────────────────────────────────

func (c *SysUserClient) TenantQuery(tenantId int64) *SysUserQuery {
	return c.Query().Where(sysuser.TenantIDEQ(tenantId), sysuser.DeletedAtIsNil())
}

func (c *SysDeptClient) TenantQuery(tenantId int64) *SysDeptQuery {
	return c.Query().Where(sysdept.TenantIDEQ(tenantId), sysdept.DeletedAtIsNil())
}

func (c *SysRoleClient) TenantQuery(tenantId int64) *SysRoleQuery {
	return c.Query().Where(sysrole.TenantIDEQ(tenantId), sysrole.DeletedAtIsNil())
}

// ─────────────────────────────────────────────
// TenantAwareQuery - 继承模式的租户查询
// 适用于 tenant_id Optional 的表（sysdict, sysdictdata）
// tenant_id = 0 表示系统默认，tenant_id = 租户ID 表示租户自定义
// 查询时返回：系统默认 + 该租户自定义 的数据
// ─────────────────────────────────────────────

func (c *SysDictClient) TenantAwareQuery(tenantId int64) *SysDictQuery {
	return c.Query().Where(
		sysdict.Or(sysdict.TenantIDEQ(tenantId), sysdict.TenantIDEQ(0)),
		sysdict.DeletedAtIsNil(),
	)
}

func (c *SysDictDataClient) TenantAwareQuery(tenantId int64) *SysDictDataQuery {
	return c.Query().Where(
		sysdictdata.Or(sysdictdata.TenantIDEQ(tenantId), sysdictdata.TenantIDEQ(0)),
		sysdictdata.DeletedAtIsNil(),
	)
}
