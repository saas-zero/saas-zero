package sysroleslogic

import (
	"context"
	"fmt"

	"github.com/saas-zero/saas-zero-common/pkg/id"
	"strconv"

	casbinapi "github.com/casbin/casbin/v2"
	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysapi"
	"github.com/saas-zero/saas-zero-basedata/ent/sysmenu"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"google.golang.org/protobuf/proto"
)

func roleToResp(r *ent.SysRole) *apps.Role {
	resp := &apps.Role{
		Id:          proto.Int64(r.ID),
		IdStr:       proto.String(id.ToString(r.ID)),
		Name:        proto.String(r.Name),
		Code:        proto.String(r.Code),
		Status:      proto.String(string(r.Status)),
		Sort:        proto.Int32(int32(r.Sort)),
		TenantId:    proto.Int64(r.TenantID),
		TenantIdStr: proto.String(id.ToString(r.TenantID)),
		CreatedAt:   proto.Int64(r.CreatedAt.UnixMilli()),
		UpdatedAt:   proto.Int64(r.UpdatedAt.UnixMilli()),
		IsSystem:    proto.Bool(r.IsSystem),
	}
	if r.Remark != "" {
		resp.Remark = proto.String(r.Remark)
	}
	if r.CreatedBy != "" {
		resp.CreatedBy = proto.String(r.CreatedBy)
	}
	if r.UpdatedBy != "" {
		resp.UpdatedBy = proto.String(r.UpdatedBy)
	}
	if len(r.Edges.Menus) > 0 {
		menuIds := make([]int64, len(r.Edges.Menus))
		for i, m := range r.Edges.Menus {
			menuIds[i] = m.ID
		}
		resp.MenuIds = menuIds
	}
	return resp
}

func roleApiIds(enf *casbinapi.SyncedEnforcer, roleCode string, tenantId int64) []int64 {
	dom := id.ToString(tenantId)
	policies, _ := enf.GetFilteredPolicy(0, roleCode, dom)
	ids := make([]int64, 0, len(policies))
	for _, p := range policies {
		if len(p) > 4 {
			if id, err := strconv.ParseInt(p[4], 10, 64); err == nil {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// currentUserAuthorized 返回当前用户"可授权范围"：已分配菜单 ID 集合、已分配 API ID 集合，
// 以及是否为 default 租户管理员（超级管理员，不设上限）。
// 用于"只能把已有权限授给别人"的继承式授权校验。
func currentUserAuthorized(svcCtx *svc.ServiceContext, ctx context.Context) (menuIDs map[int64]bool, apiIDs map[int64]bool, isDefaultAdmin bool, err error) {
	userId := mixins.GetCurrentUserId(ctx)
	tenantId := mixins.GetCurrentTenantId(ctx)

	user, uerr := svcCtx.DB.SysUser.TenantQuery(tenantId).
		Where(sysuser.IDEQ(userId)).
		WithRoles(func(q *ent.SysRoleQuery) {
			q.Where(sysrole.DeletedAtIsNil(), sysrole.StatusEQ(sysrole.StatusActive)).
				WithMenus(func(q *ent.SysMenuQuery) {
					q.Where(sysmenu.DeletedAtIsNil())
				})
		}).
		Only(ctx)
	if uerr != nil {
		return nil, nil, false, uerr
	}

	menuIDs = make(map[int64]bool)
	apiIDs = make(map[int64]bool)
	hasAdminRole := false
	for _, role := range user.Edges.Roles {
		if role.Code == "admin" {
			hasAdminRole = true
		}
		for _, menu := range role.Edges.Menus {
			menuIDs[menu.ID] = true
		}
		if svcCtx.Enforcer != nil {
			for _, apiID := range roleApiIds(svcCtx.Enforcer, role.Code, tenantId) {
				apiIDs[apiID] = true
			}
		}
	}

	// default 租户管理员（超级管理员）不限制
	if hasAdminRole && tenantId > 0 && svcCtx.DB != nil {
		tenant, terr := svcCtx.DB.SysTenant.Get(ctx, tenantId)
		if terr == nil && tenant.Code == "default" {
			return menuIDs, apiIDs, true, nil
		}
	}
	return menuIDs, apiIDs, false, nil
}

// unionMenuWithParents 将已分配菜单补全父级链，用于继承式授权范围展示/校验。
func unionMenuWithParents(svcCtx *svc.ServiceContext, ctx context.Context, menuSet map[int64]bool) (map[int64]bool, error) {
	if len(menuSet) == 0 {
		return menuSet, nil
	}
	all, err := svcCtx.DB.SysMenu.ActiveQuery().All(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]*ent.SysMenu, len(all))
	for _, m := range all {
		byID[m.ID] = m
	}
	final := make(map[int64]bool)
	var addWithParents func(id int64)
	addWithParents = func(id int64) {
		if final[id] {
			return
		}
		final[id] = true
		if m, ok := byID[id]; ok && m.ParentID > 0 {
			addWithParents(m.ParentID)
		}
	}
	for id := range menuSet {
		addWithParents(id)
	}
	return final, nil
}

// unionAPIWithGroups 将已分配 API 补全所属 group 分组，用于继承式 API 授权范围展示/校验。
func unionAPIWithGroups(svcCtx *svc.ServiceContext, ctx context.Context, apiSet map[int64]bool) (map[int64]bool, error) {
	if len(apiSet) == 0 {
		return apiSet, nil
	}
	all, err := svcCtx.DB.SysApi.ActiveQuery().All(ctx)
	if err != nil {
		return nil, err
	}
	byPath := make(map[string]*ent.SysApi)
	for _, a := range all {
		byPath[a.APIPath] = a
	}
	for _, a := range all {
		if !apiSet[a.ID] || a.APIType != sysapi.APITypeAPI {
			continue
		}
		for path, g := range byPath {
			if g.APIType == sysapi.APITypeGroup && path != "" && len(a.APIPath) > len(path) && a.APIPath[:len(path)] == path && a.APIPath[len(path)] == '/' {
				apiSet[g.ID] = true
				break
			}
		}
	}
	return apiSet, nil
}

// checkAssignableMenus 校验提交的菜单 ID 都在当前用户可授权范围内。
// default 租户管理员不受限制；其他用户只能分配自己已有的菜单（含父级补全）。
func checkAssignableMenus(svcCtx *svc.ServiceContext, ctx context.Context, menuIDs []int64) error {
	myMenus, _, isDefaultAdmin, err := currentUserAuthorized(svcCtx, ctx)
	if err != nil {
		return err
	}
	if isDefaultAdmin {
		return nil
	}
	myMenus, err = unionMenuWithParents(svcCtx, ctx, myMenus)
	if err != nil {
		return err
	}
	for _, m := range menuIDs {
		if !myMenus[m] {
			return errno.New(errno.Forbidden.Code, "无权分配菜单 ID "+id.ToString(m)+"：只能分配自己拥有的菜单")
		}
	}
	return nil
}

// checkAssignableApis 校验提交的 API ID 都在当前用户可授权范围内。
func checkAssignableApis(svcCtx *svc.ServiceContext, ctx context.Context, apiIDs []int64) error {
	_, myApis, isDefaultAdmin, err := currentUserAuthorized(svcCtx, ctx)
	if err != nil {
		return err
	}
	if isDefaultAdmin {
		return nil
	}
	myApis, err = unionAPIWithGroups(svcCtx, ctx, myApis)
	if err != nil {
		return err
	}
	for _, a := range apiIDs {
		if !myApis[a] {
			return errno.New(errno.Forbidden.Code, fmt.Sprintf("无权分配 API ID %s：只能分配自己拥有的 API", id.ToString(a)))
		}
	}
	return nil
}
