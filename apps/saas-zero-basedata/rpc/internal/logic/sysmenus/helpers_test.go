package sysmenuslogic

import (
	"testing"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysmenu"
)

func sampleMenus() []*ent.SysMenu {
	return []*ent.SysMenu{
		{ID: 1, ParentID: 0, Name: "系统管理", MenuType: sysmenu.MenuTypeDirectory, Status: sysmenu.StatusActive, Sort: 1, Path: "/system", Icon: "Setting", Hidden: false},
		{ID: 2, ParentID: 1, Name: "用户管理", MenuType: sysmenu.MenuTypeMenu, Status: sysmenu.StatusActive, Sort: 1, Path: "/system/user", Icon: "User", Hidden: false},
		{ID: 3, ParentID: 1, Name: "角色管理", MenuType: sysmenu.MenuTypeMenu, Status: sysmenu.StatusActive, Sort: 2, Path: "/system/role", Icon: "Safety", Hidden: false},
		{ID: 4, ParentID: 2, Name: "创建用户", MenuType: sysmenu.MenuTypeButton, Status: sysmenu.StatusActive, Sort: 1, Path: "", Icon: "", Hidden: false},
		{ID: 5, ParentID: 0, Name: "隐藏菜单", MenuType: sysmenu.MenuTypeMenu, Status: sysmenu.StatusActive, Sort: 99, Path: "/hidden", Hidden: true},
	}
}

func TestMenuToResp_BasicFields(t *testing.T) {
	m := &ent.SysMenu{
		ID:       100,
		Name:     "用户管理",
		MenuType: sysmenu.MenuTypeMenu,
		Status:   sysmenu.StatusActive,
		Sort:     1,
		Path:     "/system/user",
		Icon:     "UserOutlined",
		Hidden:   false,
	}

	resp := menuToResp(m)

	if resp.GetId() != 100 {
		t.Fatalf("expected id 100, got %d", resp.GetId())
	}
	if resp.GetName() != "用户管理" {
		t.Fatalf("expected name 用户管理, got %s", resp.GetName())
	}
	if resp.GetMenuType() != "menu" {
		t.Fatalf("expected menuType menu, got %s", resp.GetMenuType())
	}
	if resp.GetPath() != "/system/user" {
		t.Fatalf("expected path /system/user, got %s", resp.GetPath())
	}
	if resp.GetIcon() != "UserOutlined" {
		t.Fatalf("expected icon UserOutlined, got %s", resp.GetIcon())
	}
	if resp.GetSort() != 1 {
		t.Fatalf("expected sort 1, got %d", resp.GetSort())
	}
}

func TestBuildMenuTree(t *testing.T) {
	menus := sampleMenus()
	tree := buildMenuTree(menus, 0)

	if len(tree) != 2 {
		t.Fatalf("expected 2 root menus, got %d", len(tree))
	}

	// First root: 系统管理
	if tree[0].GetName() != "系统管理" {
		t.Fatalf("expected root menu 系统管理, got %s", tree[0].GetName())
	}
	if len(tree[0].Children) != 2 {
		t.Fatalf("expected 2 children under 系统管理, got %d", len(tree[0].Children))
	}

	// Children of 系统管理
	children := tree[0].Children
	if children[0].GetName() != "用户管理" {
		t.Fatalf("expected child 用户管理, got %s", children[0].GetName())
	}
	if children[1].GetName() != "角色管理" {
		t.Fatalf("expected child 角色管理, got %s", children[1].GetName())
	}

	// 用户管理 should have 创建用户 as child
	if len(children[0].Children) != 1 {
		t.Fatalf("expected 1 child under 用户管理, got %d", len(children[0].Children))
	}
	if children[0].Children[0].GetName() != "创建用户" {
		t.Fatalf("expected grandchild 创建用户, got %s", children[0].Children[0].GetName())
	}
}

func TestBuildRouterTree_ExcludesHiddenAndButtons(t *testing.T) {
	menus := sampleMenus()
	tree := buildRouterTree(menus, 0)

	// Should only have 系统管理 (隐藏菜单 excluded)
	if len(tree) != 1 {
		t.Fatalf("expected 1 router root menu, got %d", len(tree))
	}
	if tree[0].GetName() != "系统管理" {
		t.Fatalf("expected router root 系统管理, got %s", tree[0].GetName())
	}

	// 系统管理 should have 2 children (both menu type, not hidden)
	children := tree[0].Children
	if len(children) != 2 {
		t.Fatalf("expected 2 router children, got %d", len(children))
	}

	// 用户管理 should have NO children in router (创建用户 is a button)
	if len(children[0].Children) != 0 {
		t.Fatalf("expected 0 router children under 用户管理, got %d", len(children[0].Children))
	}
}

func TestBuildMenuTree_Empty(t *testing.T) {
	tree := buildMenuTree([]*ent.SysMenu{}, 0)
	if len(tree) != 0 {
		t.Fatal("expected empty tree for empty input")
	}
}

func TestBuildRouterTree_Empty(t *testing.T) {
	tree := buildRouterTree([]*ent.SysMenu{}, 0)
	if len(tree) != 0 {
		t.Fatal("expected empty router tree for empty input")
	}
}
