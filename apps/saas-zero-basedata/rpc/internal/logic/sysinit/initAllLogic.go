package sysinitlogic

import (
	"context"
	"strings"

	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysapi"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdept"
	"github.com/saas-zero/saas-zero-basedata/ent/sysmenu"
	"github.com/saas-zero/saas-zero-basedata/ent/syspackage"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type InitAllLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitAllLogic {
	return &InitAllLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *InitAllLogic) InitAll(_ *apps.EmptyReq) (*apps.EmptyResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)

	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	seedMenus := []struct {
		menuType  string
		name      string
		parentIdx int
		component string
		path      string
		icon      string
		sort      uint32
	}{
		{menuType: "menu", name: "控制台", parentIdx: -1, component: "dashboard/index", path: "/dashboard", icon: "Dashboard", sort: 1},
		{menuType: "directory", name: "系统管理", parentIdx: -1, path: "/system", icon: "Setting", sort: 2},
		{menuType: "menu", name: "用户管理", parentIdx: 1, component: "system/user/index", path: "/system/user", icon: "User", sort: 1},
		{menuType: "menu", name: "角色管理", parentIdx: 1, component: "system/role/index", path: "/system/role", icon: "SafetyCertificate", sort: 2},
		{menuType: "menu", name: "菜单管理", parentIdx: 1, component: "system/menu/index", path: "/system/menu", icon: "Menu", sort: 3},
		{menuType: "menu", name: "部门管理", parentIdx: 1, component: "system/dept/index", path: "/system/dept", icon: "Apartment", sort: 4},
		{menuType: "menu", name: "字典管理", parentIdx: -1, component: "dict/index", path: "/dict", icon: "Book", sort: 3},
		{menuType: "directory", name: "租户管理", parentIdx: -1, path: "/tenant", icon: "Team", sort: 4},
		{menuType: "menu", name: "租户列表", parentIdx: 7, component: "tenant/list/index", path: "/tenant/list", icon: "Team", sort: 1},
		{menuType: "menu", name: "套餐管理", parentIdx: 7, component: "tenant/package/index", path: "/tenant/package", icon: "Dollar", sort: 2},
		{menuType: "menu", name: "API管理", parentIdx: -1, component: "api/index", path: "/api", icon: "Code", sort: 5},
		{menuType: "directory", name: "日志管理", parentIdx: -1, path: "/log", icon: "FileText", sort: 6},
		{menuType: "menu", name: "登录日志", parentIdx: 11, component: "log/login-log/index", path: "/log/login-log", icon: "Login", sort: 1},
		{menuType: "menu", name: "操作日志", parentIdx: 11, component: "log/operation-log/index", path: "/log/operation-log", icon: "SwapRight", sort: 2},
		{menuType: "directory", name: "任务管理", parentIdx: -1, path: "/job", icon: "Schedule", sort: 7},
		{menuType: "menu", name: "定时任务", parentIdx: 14, component: "job/index", path: "/job/list", icon: "Schedule", sort: 1},
		{menuType: "menu", name: "任务日志", parentIdx: 14, component: "job/log/index", path: "/job/log", icon: "FileSearch", sort: 2},
		{menuType: "button", name: "用户管理", parentIdx: 2, path: "system:user:manage", sort: 1},
		{menuType: "button", name: "新增用户", parentIdx: 2, path: "system:user:create", sort: 2},
		{menuType: "button", name: "修改用户", parentIdx: 2, path: "system:user:update", sort: 3},
		{menuType: "button", name: "删除用户", parentIdx: 2, path: "system:user:delete", sort: 4},
		{menuType: "button", name: "重置密码", parentIdx: 2, path: "system:user:resetPassword", sort: 5},
		{menuType: "button", name: "分配角色", parentIdx: 2, path: "system:user:assignRoles", sort: 6},
		{menuType: "button", name: "角色管理", parentIdx: 3, path: "system:role:manage", sort: 1},
		{menuType: "button", name: "新增角色", parentIdx: 3, path: "system:role:create", sort: 2},
		{menuType: "button", name: "修改角色", parentIdx: 3, path: "system:role:update", sort: 3},
		{menuType: "button", name: "删除角色", parentIdx: 3, path: "system:role:delete", sort: 4},
		{menuType: "button", name: "分配菜单", parentIdx: 3, path: "system:role:assignMenus", sort: 5},
		{menuType: "button", name: "分配API", parentIdx: 3, path: "system:role:assignApis", sort: 6},
		{menuType: "button", name: "菜单管理", parentIdx: 4, path: "system:menu:manage", sort: 1},
		{menuType: "button", name: "新增菜单", parentIdx: 4, path: "system:menu:create", sort: 2},
		{menuType: "button", name: "修改菜单", parentIdx: 4, path: "system:menu:update", sort: 3},
		{menuType: "button", name: "删除菜单", parentIdx: 4, path: "system:menu:delete", sort: 4},
		{menuType: "button", name: "部门管理", parentIdx: 5, path: "system:dept:manage", sort: 1},
		{menuType: "button", name: "新增部门", parentIdx: 5, path: "system:dept:create", sort: 2},
		{menuType: "button", name: "修改部门", parentIdx: 5, path: "system:dept:update", sort: 3},
		{menuType: "button", name: "删除部门", parentIdx: 5, path: "system:dept:delete", sort: 4},
		{menuType: "button", name: "字典管理", parentIdx: 6, path: "system:dict:manage", sort: 1},
		{menuType: "button", name: "新增字典", parentIdx: 6, path: "system:dict:create", sort: 2},
		{menuType: "button", name: "修改字典", parentIdx: 6, path: "system:dict:update", sort: 3},
		{menuType: "button", name: "删除字典", parentIdx: 6, path: "system:dict:delete", sort: 4},
		{menuType: "button", name: "租户管理", parentIdx: 8, path: "system:tenant:manage", sort: 1},
		{menuType: "button", name: "新增租户", parentIdx: 8, path: "system:tenant:create", sort: 2},
		{menuType: "button", name: "修改租户", parentIdx: 8, path: "system:tenant:update", sort: 3},
		{menuType: "button", name: "删除租户", parentIdx: 8, path: "system:tenant:delete", sort: 4},
		{menuType: "button", name: "变更状态", parentIdx: 8, path: "system:tenant:changeStatus", sort: 5},
		{menuType: "button", name: "套餐管理", parentIdx: 9, path: "system:package:manage", sort: 1},
		{menuType: "button", name: "新增套餐", parentIdx: 9, path: "system:package:create", sort: 2},
		{menuType: "button", name: "修改套餐", parentIdx: 9, path: "system:package:update", sort: 3},
		{menuType: "button", name: "删除套餐", parentIdx: 9, path: "system:package:delete", sort: 4},
		{menuType: "button", name: "分配菜单", parentIdx: 9, path: "system:package:assignMenus", sort: 5},
		{menuType: "button", name: "分配API", parentIdx: 9, path: "system:package:assignApis", sort: 6},
		{menuType: "button", name: "API管理", parentIdx: 10, path: "system:api:manage", sort: 1},
		{menuType: "button", name: "新增API", parentIdx: 10, path: "system:api:create", sort: 2},
		{menuType: "button", name: "修改API", parentIdx: 10, path: "system:api:update", sort: 3},
		{menuType: "button", name: "删除API", parentIdx: 10, path: "system:api:delete", sort: 4},
		{menuType: "button", name: "日志查看", parentIdx: 12, path: "system:log:view", sort: 1},
		{menuType: "button", name: "任务管理", parentIdx: 15, path: "system:job:manage", sort: 1},
		{menuType: "button", name: "新增任务", parentIdx: 15, path: "system:job:create", sort: 2},
		{menuType: "button", name: "修改任务", parentIdx: 15, path: "system:job:update", sort: 3},
		{menuType: "button", name: "删除任务", parentIdx: 15, path: "system:job:delete", sort: 4},
		{menuType: "button", name: "启动任务", parentIdx: 15, path: "system:job:start", sort: 5},
		{menuType: "button", name: "暂停任务", parentIdx: 15, path: "system:job:pause", sort: 6},
		{menuType: "button", name: "立即执行", parentIdx: 15, path: "system:job:run", sort: 7},
		{menuType: "button", name: "任务日志", parentIdx: 16, path: "system:job:log", sort: 1},
	}

	tx, err := l.svcCtx.DB.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. 初始化系统默认字典。系统字典使用 tenant_id=0，重复初始化只补齐缺失项。
	if err := svc.SeedSystemDictsTx(ctx, tx); err != nil {
		return nil, err
	}

	// 2. 清除旧的通配 API（seed 时代遗留，如 /system/user/*），避免与精确接口冲突
	if _, err := tx.SysApi.Delete().Where(sysapi.APIPathContains("*")).Exec(ctx); err != nil {
		return nil, err
	}

	// 3. 创建目录 + 具体接口（幂等：已存在的跳过），并收集 /system/* 接口用于 admin 策略
	apiIds := make([]int64, 0, 80)
	adminPolicies := make([]struct{ path, method, apiId string }, 0, 60)
	for _, g := range seedApiGroups {
		// 目录（group）
		group, err := tx.SysApi.Query().Where(sysapi.APIPathEQ(g.path), sysapi.APIMethodIsNil()).First(ctx)
		if ent.IsNotFound(err) {
			group, err = tx.SysApi.Create().
				SetStatus(sysapi.StatusActive).
				SetAPIName(g.name).
				SetAPIType(sysapi.APITypeGroup).
				SetAPIPath(g.path).
				Save(ctx)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		apiIds = append(apiIds, group.ID)

		// 具体接口（api）
		for _, item := range g.apis {
			api, err := tx.SysApi.Query().Where(sysapi.APIPathEQ(item.path), sysapi.APIMethodEQ(sysapi.APIMethod(item.method))).First(ctx)
			if ent.IsNotFound(err) {
				api, err = tx.SysApi.Create().
					SetStatus(sysapi.StatusActive).
					SetAPIName(item.name).
					SetAPIType(sysapi.APITypeAPI).
					SetAPIPath(item.path).
					SetAPIMethod(sysapi.APIMethod(item.method)).
					Save(ctx)
				if err != nil {
					return nil, err
				}
			} else if err != nil {
				return nil, err
			}
			apiIds = append(apiIds, api.ID)
			// /system/* 走 basedata API 的 Casbin 校验，需为 admin 生成策略；/oauth、/init 不需要
			if strings.HasPrefix(item.path, "/system/") {
				adminPolicies = append(adminPolicies, struct{ path, method, apiId string }{
					path: item.path, method: item.method, apiId: id.ToString(api.ID),
				})
			}
		}
	}

	// 3. Create or fetch Menus (keyed by name + parent_id；button 可能重名，如多个"分配菜单")
	menuIdxToId := make(map[int]int64)
	menuIds := make([]int64, 0, len(seedMenus))
	for i, m := range seedMenus {
		var parentID int64
		if m.parentIdx >= 0 {
			parentID = menuIdxToId[m.parentIdx]
		}
		q := tx.SysMenu.Query().Where(sysmenu.NameEQ(m.name))
		if parentID > 0 {
			q = q.Where(sysmenu.ParentIDEQ(parentID))
		} else {
			q = q.Where(sysmenu.ParentIDEQ(0))
		}
		menu, err := q.First(ctx)
		if ent.IsNotFound(err) {
			create := tx.SysMenu.Create().
				SetStatus(sysmenu.StatusActive).
				SetSort(m.sort).
				SetMenuType(sysmenu.MenuType(m.menuType)).
				SetName(m.name).
				SetPath(m.path).
				SetIcon(m.icon)
			if m.component != "" {
				create.SetComponent(m.component)
			}
			if parentID > 0 {
				create.SetParentID(parentID)
			}
			menu, err = create.Save(ctx)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		menuIdxToId[i] = menu.ID
		menuIds = append(menuIds, menu.ID)
	}

	// 4. Create or fetch Package (code "standard")
	pkg, err := tx.SysPackage.Query().Where(syspackage.CodeEQ("standard")).First(ctx)
	if ent.IsNotFound(err) {
		pkg, err = tx.SysPackage.Create().
			SetName("标准套餐").
			SetCode("standard").
			SetSort(1).
			SetStatus(syspackage.StatusActive).
			AddMenuIDs(menuIds...).
			AddAPIIDs(apiIds...).
			Save(ctx)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	// 幂等：重复初始化时也确保标准套餐拥有全部菜单（含按钮）与 API，
	// 否则按套餐新建的租户会缺少按钮菜单，导致其权限码为空、前端按钮不显示。
	if err := tx.SysPackage.UpdateOneID(pkg.ID).
		ClearMenus().AddMenuIDs(menuIds...).
		ClearApis().AddAPIIDs(apiIds...).
		Exec(ctx); err != nil {
		return nil, err
	}

	// 5. Create or fetch Tenant (code "default")
	tenant, err := tx.SysTenant.Query().Where(systenant.CodeEQ("default")).First(ctx)
	if ent.IsNotFound(err) {
		tenant, err = tx.SysTenant.Create().
			SetName("默认租户").
			SetCode("default").
			SetAdminID(1).
			SetPackageID(pkg.ID).
			SetStatus(systenant.StatusActive).
			Save(ctx)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	// Update context with actual tenant ID
	ctx = mixins.SetCurrentTenantId(ctx, tenant.ID)

	// 6. Create or fetch Role (code "admin")，并确保其拥有全部菜单权限
	role, err := tx.SysRole.Query().Where(sysrole.CodeEQ("admin")).First(ctx)
	if ent.IsNotFound(err) {
		role, err = tx.SysRole.Create().
			SetName("超级管理员").
			SetCode("admin").
			SetSort(1).
			SetStatus(sysrole.StatusActive).
			AddMenuIDs(menuIds...).
			Save(ctx)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	// 幂等：重复初始化时也确保 admin 拥有全部菜单
	if err := tx.SysRole.UpdateOneID(role.ID).ClearMenus().AddMenuIDs(menuIds...).Exec(ctx); err != nil {
		return nil, err
	}

	// 7. Create or fetch Department (name "默认部门")
	dept, err := tx.SysDept.Query().Where(sysdept.NameEQ("默认部门")).First(ctx)
	if ent.IsNotFound(err) {
		dept, err = tx.SysDept.Create().
			SetName("默认部门").
			SetSort(1).
			SetStatus(sysdept.StatusActive).
			Save(ctx)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// 8. Create or fetch User (username "admin")
	_, err = tx.SysUser.Query().Where(sysuser.UsernameEQ("admin")).First(ctx)
	if ent.IsNotFound(err) {
		hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		_, err = tx.SysUser.Create().
			SetUsername("admin").
			SetPassword(string(hash)).
			SetNickname("系统管理员").
			SetDeptID(dept.ID).
			SetStatus(sysuser.StatusActive).
			AddRoleIDs(role.ID).
			Save(ctx)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 9. Casbin policies: clear old ones for admin, then re-add for all /system/* APIs
	dom := id.ToString(tenant.ID)
	if _, err := l.svcCtx.Enforcer.RemoveFilteredPolicy(0, "admin", dom); err != nil {
		logx.Errorf("initAll: failed to clear casbin policies: %v", err)
	}
	for _, p := range adminPolicies {
		if _, err := l.svcCtx.Enforcer.AddPolicy("admin", dom, p.path, strings.ToUpper(p.method), p.apiId); err != nil {
			logx.Errorf("initAll: failed to add casbin policy: %v", err)
		}
	}

	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
