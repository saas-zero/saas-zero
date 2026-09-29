package sysroleslogic

import (
	"testing"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
)

func TestRoleToResp_BasicFields(t *testing.T) {
	r := &ent.SysRole{
		ID:        100,
		Name:      "管理员",
		Code:      "admin",
		Status:    sysrole.StatusActive,
		Sort:      1,
		TenantID:  1001,
		Remark:    "系统管理员角色",
		CreatedBy: "system",
		UpdatedBy: "system",
	}

	resp := roleToResp(r)

	if resp.GetId() != r.ID {
		t.Fatalf("expected id %d, got %d", r.ID, resp.GetId())
	}
	if resp.GetIdStr() != "100" {
		t.Fatalf("expected idStr 100, got %s", resp.GetIdStr())
	}
	if resp.GetName() != "管理员" {
		t.Fatalf("expected name 管理员, got %s", resp.GetName())
	}
	if resp.GetCode() != "admin" {
		t.Fatalf("expected code admin, got %s", resp.GetCode())
	}
	if resp.GetStatus() != "active" {
		t.Fatalf("expected status active, got %s", resp.GetStatus())
	}
	if resp.GetSort() != 1 {
		t.Fatalf("expected sort 1, got %d", resp.GetSort())
	}
	if resp.GetTenantId() != 1001 {
		t.Fatalf("expected tenantId 1001, got %d", resp.GetTenantId())
	}
	if resp.GetRemark() != "系统管理员角色" {
		t.Fatalf("expected remark 系统管理员角色, got %s", resp.GetRemark())
	}
}

func TestRoleToResp_WithMenus(t *testing.T) {
	r := &ent.SysRole{
		ID: 1,
		Edges: ent.SysRoleEdges{
			Menus: []*ent.SysMenu{
				{ID: 10},
				{ID: 20},
				{ID: 30},
			},
		},
	}

	resp := roleToResp(r)

	if len(resp.GetMenuIds()) != 3 {
		t.Fatalf("expected 3 menuIds, got %d", len(resp.GetMenuIds()))
	}
	if resp.GetMenuIds()[0] != 10 || resp.GetMenuIds()[1] != 20 || resp.GetMenuIds()[2] != 30 {
		t.Fatal("menuIds mismatch")
	}
}

func TestRoleToResp_NoMenus(t *testing.T) {
	r := &ent.SysRole{ID: 1}
	resp := roleToResp(r)

	if len(resp.GetMenuIds()) != 0 {
		t.Fatalf("expected 0 menuIds for role with no menus, got %d", len(resp.GetMenuIds()))
	}
}

func TestRoleToResp_OptionalFieldsEmpty(t *testing.T) {
	r := &ent.SysRole{ID: 1}
	resp := roleToResp(r)

	if resp.Remark != nil {
		t.Fatal("expected remark to be nil when empty")
	}
	if resp.CreatedBy != nil {
		t.Fatal("expected createdBy to be nil when empty")
	}
}

func TestRoleApiIds_EmptyPolicy(t *testing.T) {
	// roleApiIds requires a real Casbin enforcer, which needs a database connection.
	// For unit tests without casbin, this function should handle nil enforcer gracefully.
	// Note: the current implementation calls enf.GetFilteredPolicy which would panic on nil.
	// This test documents that the function requires a non-nil enforcer.
}
