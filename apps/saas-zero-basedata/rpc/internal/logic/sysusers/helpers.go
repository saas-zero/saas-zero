package sysuserslogic

import (
	"context"
	"errors"

	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/logic/tenantcheck"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"google.golang.org/protobuf/proto"
)

// checkResetTarget 校验被重置密码的用户是否允许由当前操作者重置。
// 规则：
//   - default 租户 admin（系统超级管理员）只能由 default 租户管理员重置；
//   - 其他用户仅需属于当前租户（TenantQuery 已保证）即可被同租户管理员重置。
func checkResetTarget(svcCtx *svc.ServiceContext, ctx context.Context, target *ent.SysUser) error {
	tenant, err := svcCtx.DB.SysTenant.Query().
		Where(systenant.IDEQ(target.TenantID), systenant.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errno.UserNotFound
		}
		return err
	}
	if tenant.Code != "default" {
		return nil
	}
	hasAdmin, err := svcCtx.DB.SysUser.Query().
		Where(sysuser.IDEQ(target.ID), sysuser.DeletedAtIsNil()).
		QueryRoles().
		Where(sysrole.CodeEQ("admin"), sysrole.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if hasAdmin {
		return tenantcheck.RequireDefaultTenantAdmin(svcCtx, ctx)
	}
	return nil
}

// checkUserRolesInTenant 校验给定角色 ID 均属于当前租户，返回错误时禁止跨租户关联。
func checkUserRolesInTenant(svcCtx *svc.ServiceContext, ctx context.Context, tenantId int64, roleIds []int64) error {
	if len(roleIds) == 0 {
		return nil
	}
	count, err := svcCtx.DB.SysRole.TenantQuery(tenantId).
		Where(sysrole.IDIn(roleIds...)).
		Count(ctx)
	if err != nil {
		return err
	}
	if int(count) != len(roleIds) {
		return errors.New("存在不属于当前租户的角色")
	}
	return nil
}

func userToResp(u *ent.SysUser) *apps.User {
	resp := &apps.User{
		Id:          proto.Int64(u.ID),
		IdStr:       proto.String(id.ToString(u.ID)),
		Username:    proto.String(u.Username),
		Nickname:    proto.String(u.Nickname),
		Mobile:      proto.String(u.Mobile),
		Email:       proto.String(u.Email),
		Status:      proto.String(string(u.Status)),
		LoginIp:     proto.String(u.LoginIP),
		TenantId:    proto.Int64(u.TenantID),
		TenantIdStr: proto.String(id.ToString(u.TenantID)),
		CreatedAt:   proto.Int64(u.CreatedAt.UnixMilli()),
		UpdatedAt:   proto.Int64(u.UpdatedAt.UnixMilli()),
	}
	if u.CreatedBy != "" {
		resp.CreatedBy = proto.String(u.CreatedBy)
	}
	if u.UpdatedBy != "" {
		resp.UpdatedBy = proto.String(u.UpdatedBy)
	}
	if u.Remark != "" {
		resp.Remark = proto.String(u.Remark)
	}
	// SECURITY: Password is intentionally excluded from generic gRPC responses.
	// The bcrypt hash should never leak through list or detail queries.
	// Only GetUserByUsername (login flow) sets Password separately after this function returns.
	// See: getUserByUsernameLogic.go
	if u.DeptID > 0 {
		resp.DeptId = proto.Int64(u.DeptID)
		resp.DeptIdStr = proto.String(id.ToString(u.DeptID))
	}
	// 部门名称（需查询时 WithSysDept 加载，否则为空）
	if u.Edges.SysDept != nil {
		resp.DeptName = proto.String(u.Edges.SysDept.Name)
	}
	if !u.LoginAt.IsZero() {
		resp.LoginAt = proto.Int64(u.LoginAt.UnixMilli())
	}
	if len(u.Edges.Roles) > 0 {
		roleIds := make([]int64, len(u.Edges.Roles))
		roleCodes := make([]string, len(u.Edges.Roles))
		roleNames := make([]string, len(u.Edges.Roles))
		for i, r := range u.Edges.Roles {
			roleIds[i] = r.ID
			roleCodes[i] = r.Code
			roleNames[i] = r.Name
		}
		resp.RoleIds = roleIds
		resp.RoleCodes = roleCodes
		resp.RoleNames = roleNames
	}
	return resp
}
