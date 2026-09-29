package logic

import (
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	idutil "github.com/saas-zero/saas-zero-common/pkg/id"
)

// parseId converts a string ID (as returned by the frontend via `idStr`) into
// the int64 expected by the gRPC layer. JS loses precision for int64 (> 2^53),
// so IDs must travel as strings over JSON and only be parsed here on the wire.
// Implemented in saas-zero-common/pkg/id; kept as a package-local alias so the
// existing logic call sites don't need to change.
func parseId(s string) int64 {
	return idutil.Parse(s)
}

// parseIds converts a slice of string IDs to int64.
func parseIds(ss []string) []int64 {
	return idutil.ParseStrings(ss)
}

// toSysUser maps the gRPC User to the HTTP response type.
// id (int64) is the numeric ID, idStr (string) is for frontend precision.
func toSysUser(u *apps.User) *types.SysUser {
	if u == nil {
		return nil
	}
	return &types.SysUser{
		Id:          u.GetId(),
		IdStr:       u.GetIdStr(),
		Username:    u.GetUsername(),
		Nickname:    u.GetNickname(),
		Mobile:      u.GetMobile(),
		Email:       u.GetEmail(),
		DeptId:      u.GetDeptId(),
		DeptIdStr:   u.GetDeptIdStr(),
		DeptName:    u.GetDeptName(),
		Status:      u.GetStatus(),
		Remark:      u.GetRemark(),
		LoginIp:     u.GetLoginIp(),
		LastLoginAt: auditTime(u.GetLoginAt()),
		RoleIds:     idutil.ToStrings(u.GetRoleIds()),
		RoleCodes:   u.GetRoleCodes(),
		RoleNames:   u.GetRoleNames(),
		TenantId:    u.GetTenantId(),
		TenantIdStr: u.GetTenantIdStr(),
		CreatedAt:   auditTime(u.GetCreatedAt()),
		CreatedBy:   u.GetCreatedBy(),
		UpdatedAt:   auditTime(u.GetUpdatedAt()),
		UpdatedBy:   u.GetUpdatedBy(),
	}
}

// toSysUserList maps a slice of gRPC users.
func toSysUserList(users []*apps.User) []*types.SysUser {
	out := make([]*types.SysUser, 0, len(users))
	for _, u := range users {
		out = append(out, toSysUser(u))
	}
	return out
}
