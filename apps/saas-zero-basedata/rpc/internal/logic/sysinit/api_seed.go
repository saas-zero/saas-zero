package sysinitlogic

// seedApiGroup 定义一组 API 目录及其下的具体接口。
// 目录（group）只用于前端分组展示，不生成 Casbin 策略；
// 具体接口（api）精确到 path + method，授权后写入 Casbin 策略。
type seedApiGroup struct {
	name string // 目录名称
	path string // 目录路径（路径前缀）
	apis []seedApiItem
}

type seedApiItem struct {
	name   string // 接口名称
	path   string // 接口路径
	method string // HTTP 方法（小写，与 sys_api 枚举一致）
}

// seedApiGroups 是全部服务的接口清单（数据源：gateway.yaml / 各服务路由注册）。
// 已核对：system 56 个 + oauth 9 个 + init 5 个，共 70 个接口。
var seedApiGroups = []seedApiGroup{
	{
		name: "用户管理",
		path: "/system/user",
		apis: []seedApiItem{
			{name: "新增用户", path: "/system/user/create", method: "post"},
			{name: "更新用户", path: "/system/user/update", method: "post"},
			{name: "删除用户", path: "/system/user/delete", method: "post"},
			{name: "用户列表", path: "/system/user/list", method: "get"},
			{name: "用户详情", path: "/system/user/detail", method: "get"},
			{name: "分配角色", path: "/system/user/assignRoles", method: "post"},
			{name: "重置密码", path: "/system/user/resetPassword", method: "post"},
		},
	},
	{
		name: "角色管理",
		path: "/system/role",
		apis: []seedApiItem{
			{name: "新增角色", path: "/system/role/create", method: "post"},
			{name: "更新角色", path: "/system/role/update", method: "post"},
			{name: "删除角色", path: "/system/role/delete", method: "post"},
			{name: "角色列表", path: "/system/role/list", method: "get"},
			{name: "角色详情", path: "/system/role/detail", method: "get"},
			{name: "分配菜单", path: "/system/role/assignMenus", method: "post"},
			{name: "分配API", path: "/system/role/assignApis", method: "post"},
		},
	},
	{
		name: "菜单管理",
		path: "/system/menu",
		apis: []seedApiItem{
			{name: "新增菜单", path: "/system/menu/create", method: "post"},
			{name: "更新菜单", path: "/system/menu/update", method: "post"},
			{name: "删除菜单", path: "/system/menu/delete", method: "post"},
			{name: "菜单列表", path: "/system/menu/list", method: "get"},
			{name: "菜单详情", path: "/system/menu/detail", method: "get"},
			{name: "菜单树", path: "/system/menu/tree", method: "get"},
			{name: "菜单路由", path: "/system/menu/routers", method: "get"},
		},
	},
	{
		name: "部门管理",
		path: "/system/dept",
		apis: []seedApiItem{
			{name: "新增部门", path: "/system/dept/create", method: "post"},
			{name: "更新部门", path: "/system/dept/update", method: "post"},
			{name: "删除部门", path: "/system/dept/delete", method: "post"},
			{name: "部门列表", path: "/system/dept/list", method: "get"},
			{name: "部门详情", path: "/system/dept/detail", method: "get"},
			{name: "部门树", path: "/system/dept/tree", method: "get"},
		},
	},
	{
		name: "字典管理",
		path: "/system/dict",
		apis: []seedApiItem{
			{name: "新增字典", path: "/system/dict/create", method: "post"},
			{name: "更新字典", path: "/system/dict/update", method: "post"},
			{name: "删除字典", path: "/system/dict/delete", method: "post"},
			{name: "字典列表", path: "/system/dict/list", method: "get"},
			{name: "字典详情", path: "/system/dict/detail", method: "get"},
		},
	},
	{
		name: "字典数据管理",
		path: "/system/dictData",
		apis: []seedApiItem{
			{name: "新增字典数据", path: "/system/dictData/create", method: "post"},
			{name: "更新字典数据", path: "/system/dictData/update", method: "post"},
			{name: "删除字典数据", path: "/system/dictData/delete", method: "post"},
			{name: "字典数据列表", path: "/system/dictData/list", method: "get"},
			{name: "字典数据详情", path: "/system/dictData/detail", method: "get"},
			{name: "按Key查询", path: "/system/dictData/byDictKey", method: "get"},
		},
	},
	{
		name: "租户管理",
		path: "/system/tenant",
		apis: []seedApiItem{
			{name: "新增租户", path: "/system/tenant/create", method: "post"},
			{name: "更新租户", path: "/system/tenant/update", method: "post"},
			{name: "删除租户", path: "/system/tenant/delete", method: "post"},
			{name: "租户列表", path: "/system/tenant/list", method: "get"},
			{name: "租户详情", path: "/system/tenant/detail", method: "get"},
			{name: "变更租户状态", path: "/system/tenant/changeStatus", method: "post"},
			{name: "租户用户列表", path: "/system/tenant/users", method: "get"},
		},
	},
	{
		name: "套餐管理",
		path: "/system/package",
		apis: []seedApiItem{
			{name: "新增套餐", path: "/system/package/create", method: "post"},
			{name: "更新套餐", path: "/system/package/update", method: "post"},
			{name: "删除套餐", path: "/system/package/delete", method: "post"},
			{name: "套餐列表", path: "/system/package/list", method: "get"},
			{name: "套餐详情", path: "/system/package/detail", method: "get"},
			{name: "分配套餐菜单", path: "/system/package/assignMenus", method: "post"},
			{name: "分配套餐API", path: "/system/package/assignApis", method: "post"},
		},
	},
	{
		name: "API管理",
		path: "/system/api",
		apis: []seedApiItem{
			{name: "新增API", path: "/system/api/create", method: "post"},
			{name: "更新API", path: "/system/api/update", method: "post"},
			{name: "删除API", path: "/system/api/delete", method: "post"},
			{name: "API列表", path: "/system/api/list", method: "get"},
			{name: "我的API", path: "/system/api/mine", method: "get"},
			{name: "API详情", path: "/system/api/detail", method: "get"},
		},
	},
	{
		name: "日志管理",
		path: "/system/log",
		apis: []seedApiItem{
			{name: "登录日志", path: "/system/log/loginLog/list", method: "get"},
			{name: "操作日志", path: "/system/log/operationLog/list", method: "get"},
		},
	},
	{
		name: "任务管理",
		path: "/system/job",
		apis: []seedApiItem{
			{name: "新增任务", path: "/system/job/create", method: "post"},
			{name: "更新任务", path: "/system/job/update", method: "post"},
			{name: "删除任务", path: "/system/job/delete", method: "post"},
			{name: "任务列表", path: "/system/job/list", method: "get"},
			{name: "任务详情", path: "/system/job/detail", method: "get"},
			{name: "启动任务", path: "/system/job/start", method: "post"},
			{name: "暂停任务", path: "/system/job/pause", method: "post"},
			{name: "立即执行", path: "/system/job/runOnce", method: "post"},
			{name: "任务日志", path: "/system/job/log/list", method: "get"},
			{name: "清理日志", path: "/system/job/log/clean", method: "post"},
		},
	},
}
