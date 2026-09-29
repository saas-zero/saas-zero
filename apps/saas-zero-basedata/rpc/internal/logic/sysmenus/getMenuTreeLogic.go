package sysmenuslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysmenu"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuTreeLogic {
	return &GetMenuTreeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMenuTreeLogic) GetMenuTree(_ *apps.EmptyReq) (*apps.MenuTreeResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	isDefaultTenantAdmin := false
	var allMenus []*ent.SysMenu
	var err error

	// 仅 default 租户的管理员（超级管理员）拥有全量菜单；其他用户一律按角色菜单并集返回
	if userId > 0 {
		user, uerr := l.svcCtx.DB.SysUser.ActiveQuery().
			Where(sysuser.IDEQ(userId)).
			WithRoles(func(q *ent.SysRoleQuery) {
				q.Where(sysrole.DeletedAtIsNil(), sysrole.StatusEQ(sysrole.StatusActive)).
					WithMenus(func(q *ent.SysMenuQuery) {
						q.Where(sysmenu.DeletedAtIsNil())
					})
			}).
			Only(l.ctx)
		if uerr == nil {
			hasAdminRole := false
			menuSet := make(map[int64]bool)
			for _, role := range user.Edges.Roles {
				if role.Code == "admin" {
					hasAdminRole = true
				}
				for _, menu := range role.Edges.Menus {
					menuSet[menu.ID] = true
				}
			}

			// 判断当前租户是否是 default 租户
			if hasAdminRole && tenantId > 0 {
				tenant, terr := l.svcCtx.DB.SysTenant.Get(l.ctx, tenantId)
				if terr == nil && tenant.Code == "default" {
					isDefaultTenantAdmin = true
				}
			}

			if isDefaultTenantAdmin {
				allMenus, err = l.svcCtx.DB.SysMenu.ActiveQuery().
					Order(ent.Asc(sysmenu.FieldSort)).
					All(l.ctx)
			} else {
				allMenus, err = unionMenusWithParents(l.svcCtx, l.ctx, menuSet)
			}
		}
	}

	if err != nil {
		return nil, err
	}

	tree := buildMenuTree(allMenus, 0)
	return &apps.MenuTreeResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: tree,
	}, nil
}

// unionMenusWithParents 返回角色分配的菜单并集，并补全其父级链，
// 避免角色只分配了子菜单时前端树出现悬空节点（父菜单缺失导致子菜单不可见）。
func unionMenusWithParents(svcCtx *svc.ServiceContext, ctx context.Context, menuSet map[int64]bool) ([]*ent.SysMenu, error) {
	if len(menuSet) == 0 {
		return nil, nil
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
	ids := make([]int64, 0, len(final))
	for id := range final {
		ids = append(ids, id)
	}
	return svcCtx.DB.SysMenu.ActiveQuery().
		Where(sysmenu.IDIn(ids...)).
		Order(ent.Asc(sysmenu.FieldSort)).
		All(ctx)
}
