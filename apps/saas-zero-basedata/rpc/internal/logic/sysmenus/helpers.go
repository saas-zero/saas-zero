package sysmenuslogic

import (
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysmenu"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"google.golang.org/protobuf/proto"
)

func menuToResp(m *ent.SysMenu) *apps.Menu {
	resp := &apps.Menu{
		Id:         proto.Int64(m.ID),
		IdStr:      proto.String(id.ToString(m.ID)),
		MenuType:   proto.String(string(m.MenuType)),
		Name:       proto.String(m.Name),
		Component:  proto.String(m.Component),
		Path:       proto.String(m.Path),
		Icon:       proto.String(m.Icon),
		IsRedirect: proto.Bool(m.IsRedirect),
		Redirect:   proto.String(m.Redirect),
		Hidden:     proto.Bool(m.Hidden),
		Status:     proto.String(string(m.Status)),
		Sort:       proto.Int32(int32(m.Sort)),
		Remark:     proto.String(m.Remark),
		CreatedAt:  proto.Int64(m.CreatedAt.UnixMilli()),
		UpdatedAt:  proto.Int64(m.UpdatedAt.UnixMilli()),
	}
	if m.CreatedBy != "" {
		resp.CreatedBy = proto.String(m.CreatedBy)
	}
	if m.UpdatedBy != "" {
		resp.UpdatedBy = proto.String(m.UpdatedBy)
	}
	if m.ParentID > 0 {
		resp.ParentId = proto.Int64(m.ParentID)
		resp.ParentIdStr = proto.String(id.ToString(m.ParentID))
	}
	return resp
}

func buildMenuTree(menus []*ent.SysMenu, parentId int64) []*apps.Menu {
	var result []*apps.Menu
	for _, m := range menus {
		if m.ParentID == parentId {
			item := menuToResp(m)
			item.Children = buildMenuTree(menus, m.ID)
			result = append(result, item)
		}
	}
	return result
}

func buildRouterTree(menus []*ent.SysMenu, parentId int64) []*apps.Menu {
	// Routers: only visible menus (not hidden, active, directory/menu types)
	var result []*apps.Menu
	for _, m := range menus {
		if m.ParentID == parentId && !m.Hidden && m.Status == sysmenu.StatusActive && m.MenuType != sysmenu.MenuTypeButton {
			item := menuToResp(m)
			item.Children = buildRouterTree(menus, m.ID)
			result = append(result, item)
		}
	}
	return result
}
