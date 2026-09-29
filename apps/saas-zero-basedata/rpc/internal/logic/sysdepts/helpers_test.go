package sysdeptslogic

import (
	"testing"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdept"
)

func TestDeptToResp_BasicFields(t *testing.T) {
	d := &ent.SysDept{
		ID:       100,
		Name:     "技术部",
		Status:   sysdept.StatusActive,
		Sort:     1,
		Mobile:   "010-12345678",
		Email:    "tech@example.com",
		TenantID: 1001,
		ParentID: 50,
		LeaderID: 200,
	}

	resp := deptToResp(d)

	if resp.GetId() != 100 {
		t.Fatalf("expected id 100, got %d", resp.GetId())
	}
	if resp.GetName() != "技术部" {
		t.Fatalf("expected name 技术部, got %s", resp.GetName())
	}
	if resp.GetMobile() != "010-12345678" {
		t.Fatalf("expected mobile 010-12345678, got %s", resp.GetMobile())
	}
	if resp.GetEmail() != "tech@example.com" {
		t.Fatalf("expected email tech@example.com, got %s", resp.GetEmail())
	}
	if resp.GetParentId() != 50 {
		t.Fatalf("expected parentId 50, got %d", resp.GetParentId())
	}
	if resp.GetLeaderId() != 200 {
		t.Fatalf("expected leaderId 200, got %d", resp.GetLeaderId())
	}
}

func TestDeptToResp_WithLeader(t *testing.T) {
	d := &ent.SysDept{
		ID: 1,
		Edges: ent.SysDeptEdges{
			Leader: &ent.SysUser{
				Nickname: "张三",
			},
		},
	}

	resp := deptToResp(d)
	if resp.GetLeaderName() != "张三" {
		t.Fatalf("expected leaderName 张三, got %s", resp.GetLeaderName())
	}
}

func TestDeptToResp_OptionalFieldsEmpty(t *testing.T) {
	d := &ent.SysDept{ID: 1}
	resp := deptToResp(d)

	if resp.ParentId != nil {
		t.Fatal("expected parentId nil when not set")
	}
	if resp.LeaderId != nil {
		t.Fatal("expected leaderId nil when not set")
	}
	if resp.LeaderName != nil {
		t.Fatal("expected leaderName nil when no leader edges")
	}
}

func TestBuildDeptTree(t *testing.T) {
	depts := []*ent.SysDept{
		{ID: 1, ParentID: 0, Name: "总公司", Sort: 1},
		{ID: 2, ParentID: 1, Name: "技术部", Sort: 1},
		{ID: 3, ParentID: 1, Name: "市场部", Sort: 2},
		{ID: 4, ParentID: 2, Name: "前端组", Sort: 1},
		{ID: 5, ParentID: 0, Name: "分公司", Sort: 2},
	}

	tree := buildDeptTree(depts, 0)

	if len(tree) != 2 {
		t.Fatalf("expected 2 root departments, got %d", len(tree))
	}
	if tree[0].GetName() != "总公司" {
		t.Fatalf("expected root 总公司, got %s", tree[0].GetName())
	}

	children := tree[0].Children
	if len(children) != 2 {
		t.Fatalf("expected 2 children under 总公司, got %d", len(children))
	}
	if children[0].GetName() != "技术部" {
		t.Fatalf("expected child 技术部, got %s", children[0].GetName())
	}

	// 技术部 should have 前端组 as child
	if len(children[0].Children) != 1 {
		t.Fatalf("expected 1 child under 技术部, got %d", len(children[0].Children))
	}
	if children[0].Children[0].GetName() != "前端组" {
		t.Fatalf("expected grandchild 前端组, got %s", children[0].Children[0].GetName())
	}
}
