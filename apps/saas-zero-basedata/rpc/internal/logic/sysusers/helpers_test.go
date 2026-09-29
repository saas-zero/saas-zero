package sysuserslogic

import (
	"testing"
	"time"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
)

func TestUserToResp_BasicFields(t *testing.T) {
	now := time.Now()
	u := &ent.SysUser{
		ID:        123456789012345678,
		Username:  "admin",
		Nickname:  "管理员",
		Mobile:    "13800138000",
		Email:     "admin@example.com",
		Status:    sysuser.StatusActive,
		LoginIP:   "192.168.1.1",
		TenantID:  1001,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: "system",
		UpdatedBy: "system",
		Remark:    "测试用户",
	}
	u.LoginAt = now

	resp := userToResp(u)

	if resp.GetId() != u.ID {
		t.Fatalf("expected id %d, got %d", u.ID, resp.GetId())
	}
	if resp.GetIdStr() != "123456789012345678" {
		t.Fatalf("expected idStr 123456789012345678, got %s", resp.GetIdStr())
	}
	if resp.GetUsername() != "admin" {
		t.Fatalf("expected username admin, got %s", resp.GetUsername())
	}
	if resp.GetNickname() != "管理员" {
		t.Fatalf("expected nickname 管理员, got %s", resp.GetNickname())
	}
	if resp.GetMobile() != "13800138000" {
		t.Fatalf("expected mobile 13800138000, got %s", resp.GetMobile())
	}
	if resp.GetEmail() != "admin@example.com" {
		t.Fatalf("expected email admin@example.com, got %s", resp.GetEmail())
	}
	if resp.GetStatus() != "active" {
		t.Fatalf("expected status active, got %s", resp.GetStatus())
	}
	if resp.GetLoginIp() != "192.168.1.1" {
		t.Fatalf("expected loginIp 192.168.1.1, got %s", resp.GetLoginIp())
	}
	if resp.GetTenantId() != 1001 {
		t.Fatalf("expected tenantId 1001, got %d", resp.GetTenantId())
	}
	if resp.GetRemark() != "测试用户" {
		t.Fatalf("expected remark 测试用户, got %s", resp.GetRemark())
	}
}

func TestUserToResp_PasswordExcluded(t *testing.T) {
	// Password should NOT be set by userToResp for security reasons.
	// It is only set separately by GetUserByUsername.
	u := &ent.SysUser{
		ID:       1,
		Username: "test",
		Password: "$2a$10$supersecrethashvalue1234567890",
	}

	resp := userToResp(u)
	if resp.Password != nil {
		t.Fatal("password should be nil in generic userToResp response")
	}
}

func TestUserToResp_DepartmentID(t *testing.T) {
	u := &ent.SysUser{
		ID:     1,
		DeptID: 5001,
	}

	resp := userToResp(u)
	if resp.GetDeptId() != 5001 {
		t.Fatalf("expected deptId 5001, got %d", resp.GetDeptId())
	}
	if resp.GetDeptIdStr() != "5001" {
		t.Fatalf("expected deptIdStr 5001, got %s", resp.GetDeptIdStr())
	}
}

func TestUserToResp_WithRoles(t *testing.T) {
	u := &ent.SysUser{
		ID: 1,
		Edges: ent.SysUserEdges{
			Roles: []*ent.SysRole{
				{ID: 10, Code: "admin", Name: "管理员"},
				{ID: 20, Code: "user", Name: "普通用户"},
			},
		},
	}

	resp := userToResp(u)

	if len(resp.GetRoleIds()) != 2 {
		t.Fatalf("expected 2 roleIds, got %d", len(resp.GetRoleIds()))
	}
	if resp.GetRoleIds()[0] != 10 {
		t.Fatalf("expected roleId[0]=10, got %d", resp.GetRoleIds()[0])
	}
	if len(resp.GetRoleCodes()) != 2 {
		t.Fatalf("expected 2 roleCodes, got %d", len(resp.GetRoleCodes()))
	}
	if resp.GetRoleCodes()[0] != "admin" {
		t.Fatalf("expected roleCodes[0]=admin, got %s", resp.GetRoleCodes()[0])
	}
	if resp.GetRoleNames()[0] != "管理员" {
		t.Fatalf("expected roleNames[0]=管理员, got %s", resp.GetRoleNames()[0])
	}
}

func TestUserToResp_NoRoles(t *testing.T) {
	u := &ent.SysUser{ID: 1}
	resp := userToResp(u)

	if len(resp.GetRoleIds()) != 0 {
		t.Fatalf("expected 0 roleIds for user with no roles, got %d", len(resp.GetRoleIds()))
	}
	if len(resp.GetRoleCodes()) != 0 {
		t.Fatalf("expected 0 roleCodes, got %d", len(resp.GetRoleCodes()))
	}
}

func TestUserToResp_OptionalFieldsEmpty(t *testing.T) {
	u := &ent.SysUser{ID: 1}

	resp := userToResp(u)

	if resp.Remark != nil {
		t.Fatal("expected remark to be nil when empty")
	}
	if resp.DeptId != nil {
		t.Fatal("expected deptId to be nil when not set")
	}
	if resp.LoginAt != nil {
		t.Fatal("expected loginAt to be nil for zero time")
	}
}
