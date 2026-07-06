# SaaS-Zero 架构设计

## 一、总体架构

### 分层设计

```
                    ┌─────────────┐
                    │  客户端/前端   │
                    └──────┬──────┘
                           │ HTTP
                    ┌──────▼──────┐
                    │   API 网关   │ ← saas-zero-gateway (go-zero gateway)
                    │ (HTTP 纯代理) │  只做路由映射，不解析 JWT，不做 Casbin
                    └──┬──────┬───┘
                       │      │
                HTTP   │      │ HTTP
                 ┌─────▼──┐ ┌─▼──────────────┐
                 │ 认证服务 │ │ 基础数据 API    │ ← saas-zero-auth    (HTTP :18081)
                 │ (auth)  │ │ saas-zero-basedata│ ← saas-zero-basedata (HTTP :18083)
                 └────┬────┘ │ ┌──────────────┤
                      │ gRPC │ │ JWT 中间件    │  ← 解析 token → context
                      │      │ │ Casbin 中间件  │  ← 权限检查
                      │      │ │ Logic → gRPC  │
                      │      └─┴──────┬───────┘
                      │              │ gRPC
                      │       ┌──────▼──────┐
                      └───────│ 基础数据 RPC  │ ← saas-zero-basedata (gRPC :18084)
                              │ Ent DB 操作   │
                              └──────┬──────┘
                                     │ Ent
                              ┌──────▼──────┐
                              │  PostgreSQL  │
                              └─────────────┘
```

### 服务职责

| 服务 | 协议 | 端口 | 职责 |
|---|---|---|---|
| **gateway** | HTTP 代理 | `:18080` | 统一入口，按路径转发到对应 upstream，不做任何鉴权 |
| **auth** | HTTP API + gRPC client | `:18081` | 登录认证、JWT 签发、令牌验证/刷新、用户信息/菜单/权限查询 |
| **basedata API** | HTTP API + gRPC client | `:18083` | 对外暴露 CRUD 接口、JWT 解析注入、Casbin 运行时权限校验、调 RPC |
| **basedata RPC** | gRPC server | `:18084` | 核心业务逻辑、Ent DB 操作、Casbin 策略管理、自动审计字段填充 |

**设计原则：** Gateway 是唯一对外暴露的节点，auth 和 basedata API 只对 gateway 开放（内网访问）。

### 通信方式

- **外部 → Gateway：** HTTP（RESTful），go-zero `gateway` 包，纯 HTTP 代理转发
- **Gateway → Auth：** HTTP 代理（`/oauth/*` 路径）
- **Gateway → Basedata API：** HTTP 代理（`/system/*`、`/init/*` 路径）
- **Auth → Basedata RPC：** gRPC，auth 在登录/查询时需要操作数据库
- **Basedata API → Basedata RPC：** gRPC，API 层逻辑调 RPC 层进行 DB 操作

## 二、多租户设计

### 租户隔离模型

```
┌──────────────────────────────────────────────┐
│                 SaaS 平台                      │
├──────────────┬───────────────────────────────┤
│  系统租户     │         租户 A   租户 B   租户 C │
│  (平台本身)   │         ┌───┐  ┌───┐  ┌───┐  │
│              │         │用 │  │用 │  │用 │  │
│              │         │户 │  │户 │  │户 │  │
│              │         │角 │  │角 │  │角 │  │
│              │         │色 │  │色 │  │色 │  │
│              │         │菜 │  │菜 │  │菜 │  │
│              │         │单 │  │单 │  │单 │  │
│              │         └───┘  └───┘  └───┘  │
└──────────────┴───────────────────────────────┘
```

**隔离级别：** 采用 **共享数据库 + 逻辑隔离** 模型。所有租户共享同一数据库实例，通过 `tenant_id` 列进行逻辑隔离。

**Casbin 多租户：** 采用 Domain RBAC 模型，将 `tenant_id` 作为 domain 参数传入，实现运行时 API 权限的租户隔离。

### 数据隔离实现

| 模式 | 说明 | 适用表 |
|---|---|---|
| **必填隔离** | `tenant_id > 0`，每条数据必须属于某个租户 | sys_user, sys_role, sys_dept |
| **继承隔离** | `tenant_id >= 0`，0=系统默认，>0=租户自定义 | sys_dict, sys_dict_data |
| **无隔离** | 无 tenant_id 字段，全局共享 | sys_tenant, sys_menu, sys_api, sys_package, sys_login_log, sys_operation_log |

> 注意：sys_menu 和 sys_tenant 实际无 tenant_id 字段（sys_menus、sys_tenants 表无 TenantMixin），属于全局共享数据。sys_package 同样无 tenant_id。

#### 查询辅助方法

为防止开发者遗漏 `tenant_id` 过滤，每个 Client 提供了标准化的查询入口：

```go
// 有租户隔离 + 软删除
users, err := client.SysUser.TenantQuery(tenantId).All(ctx)

// 字典继承模式（返回系统默认 + 租户自定义）
dicts, err := client.SysDict.TenantAwareQuery(tenantId).All(ctx)

// 无租户隔离（如系统日志、API目录）
apis, err := client.SysApi.Query().All(ctx)
```

### 字典继承机制

```
字典类型: "order_status"
├── 系统默认 (tenant_id=0)
│   ├── pending    → "待支付"
│   ├── paid       → "已支付"
│   └── shipped    → "已发货"
└── 租户 A 自定义 (tenant_id=1001)
    └── pending    → "待付款"  ← 覆盖系统默认
```

**查询 SQL：** `WHERE (tenant_id = ? OR tenant_id = 0) AND deleted_at IS NULL`

**优先级：** 应用层处理——查询时一起返回，Logic 层按"租户自定义覆盖系统默认"合并。

### 租户信息的传递

```
请求 → Basedata API JWT 中间件 (解析 JWT → 提取 tenant_id/username/userId)
     → context.WithValue (mixins.SetCurrentTenantId/UserId/UserName)
     → gRPC Metadata (附加 x-user-id, x-user-name, x-tenant-id)
     → Basedata RPC Auth Interceptor (从 metadata 读取)
     → mixins.SetCurrent* (写入 context)
     → Ent Client (Mixin Hook 自动写入 DB)
```

## 三、认证与授权

### 认证流程

```
1. 用户 → POST /oauth/login → Auth 服务
2. Auth → gRPC GetUserByUsername → 查用户
3. Auth → bcrypt.Verify 密码
4. Auth → 生成 JWT（含 userId, tenantId, userName, roleCodes）
5. 返回 JWT 给前端
```

### 授权数据流

```
请求 → Gateway → Basedata API
                  │
                  ├─ JWT 中间件 (所有 /system/* 路由)
                  │   ├─ Bearer token → jwt.Parse → {userId, tenantId, userName, roleCodes}
                  │   └─ context = mixins.SetCurrentUserId/Name/TenantId(ctx, ...)
                  │      context = context.WithValue(ctx, "role_codes", claims.RoleCodes)
                  │
                  ├─ Casbin 中间件 (跳过 /init/*)
                  │   ├─ roleCodes = GetRoleCodes(ctx)
                  │   ├─ tenantId = mixins.GetCurrentTenantId(ctx)
                  │   ├─ dom = strconv.FormatInt(tenantId, 10)
                  │   ├─ for each roleCode: enforcer.Enforce(roleCode, dom, path, method)
                  │   └─ 全拒 → 403 Forbidden
                  │
                  └─ Logic → gRPC (context 已有 auth info)
                              └─ gRPC Auth Interceptor → mixins.SetCurrent*()
                                  └─ Ent Mixin Hook → 自动填充审计字段
```

### JWT Claims

```go
type Claims struct {
    UserId    int64    `json:"userId"`
    TenantId  int64    `json:"tenantId"`
    UserName  string   `json:"userName"`
    RoleCodes []string `json:"roleCodes"`  // 登录时从 user.GetRoleCodes() 写入
    gojwt.RegisteredClaims
}
```

### Casbin 策略模型

```
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, ept

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.dom == p.dom && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)
```

- `sub` = roleCode（如 "admin"），`dom` = tenantId 字符串（如 "1001"）
- `obj` = API 路径，支持 keyMatch 通配 `*`（如 `/system/user/*`）
- `act` = HTTP 方法，支持正则（如 `POST`、`GET|POST`）
- `ept` = apiId（额外存储，不参与匹配，用于反向查询）

### 策略存储

```sql
CREATE TABLE casbin_rule (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL DEFAULT '',
    v0 VARCHAR(100) NOT NULL DEFAULT '',  -- roleCode
    v1 VARCHAR(100) NOT NULL DEFAULT '',  -- tenantId (dom)
    v2 VARCHAR(100) NOT NULL DEFAULT '',  -- api path
    v3 VARCHAR(100) NOT NULL DEFAULT '',  -- http method
    v4 VARCHAR(100) NOT NULL DEFAULT '',  -- apiId (extra)
    v5 VARCHAR(100) NOT NULL DEFAULT ''
);
```

- Casbin auto-save 自动写库，`AddPolicy/RemovePolicy` 立即持久化
- 两个进程（RPC 管理、API 检查）各自独立 `SyncedEnforcer` 实例
- 全量加载启动，每次策略变更实时写入 `casbin_rule` 表

### 权限数据结构

```
sys_user ──── M:N ──── sys_role ──── M:N ──── sys_menu  (前端菜单权限，通过 ent edge 管理)
                           │
                      Casbin Policy (casbin_rule 表):      (API 运行时权限)
                      p, roleCode, tenantId, path, method
                           │
                      sys_api  (API 资源目录，管理后台 UI 展示用，不影响鉴权)
```

API 鉴权和菜单权限分离：
- **API 权限**：通过 Casbin 策略管理，运行时由 middleware 拦截校验
- **菜单权限**：通过 `sys_role_menus` 关联表（ent edge）管理，前端根据角色加载菜单树

## 四、数据层设计

### Ent ORM + Mixin 体系

```
每个 Schema 由三部分组成：

┌─────────────────┐
│   业务字段        │  ← 手动定义
├─────────────────┤
│   Mixin 混入      │  ← 标准化字段
│   ├─ BaseMixin   │     id (雪花ID)
│   ├─ TenantMixin │     tenant_id (必填/可选)
│   ├─ CreatedMixin│     created_at/created_id/created_by
│   ├─ UpdatedMixin│     updated_at/updated_id/updated_by
│   ├─ DeletedMixin│     deleted_at/deleted_id/deleted_by (软删除)
│   ├─ StatusMixin │     status (active/inactive/suspended)
│   ├─ SortMixin   │     sort (排序号，可选)
│   └─ RemarkMixin │     remark (备注，可选)
├─────────────────┤
│   关联(Edges)     │  ← 表间关系定义
└─────────────────┘
```

### Mixin Hook 机制

| Mixin | 触发时机 | 自动填充字段 | 数据来源 |
|---|---|---|---|
| CreatedMixin | OpCreate | created_at, created_id, created_by | context (SetCurrentUserId/Name) |
| UpdatedMixin | OpCreate, OpUpdate, OpUpdateOne | updated_at, updated_id, updated_by | context (SetCurrentUserId/Name) |
| DeletedMixin | OpUpdate/OpUpdateOne | deleted_at, deleted_id, deleted_by | context (SetCurrentUserId/Name) |
| TenantMixin | OpCreate | tenant_id | context (SetCurrentTenantId) |

Hooks 在 mutation 执行前触发，通过 `ent.Mutation.SetField()` 设值，然后交回给 ent 标准流程执行。

### 软删除设计

- **业务状态**用 `status` 字段（active/inactive/suspended）
- **删除标记**用 `deleted_at` + `deleted_id` + `deleted_by` 字段
- 所有查询默认通过 `ActiveQuery()` / `TenantQuery()` 自动过滤已删除记录
- 需要查全部（含删除）：使用原生 `.Query()`

## 五、实体关系图

```
SysTenant
    │
    ├── SysUser (tenant_id FK)
    │       └── M:N via sys_user_roles ──── SysRole (tenant_id FK)
    │               ├── M:N via sys_role_menus ──── SysMenu (无 tenant_id)
    │               └── Casbin Policy (casbin_rule) → SysApi (无 tenant_id)
    │       └── FK ──── SysDept (tenant_id FK, leader_id FK → SysUser)
    │
    ├── SysDept (tenant_id FK, leader_id FK → SysUser)
    │       └── tree (parent_id self-ref)
    │
    ├── SysDict (tenant_id FK, Optional)
    │       └── SysDictData (tenant_id FK, Optional, dict_id FK)
    │
    ├── SysMenu (无 tenant_id)
    │       └── tree (parent_id self-ref)
    │
    ├── SysPackage (无 tenant_id)
    │       ├── M:N via sys_package_menus ──── SysMenu
    │       └── M:N via sys_package_apis ──── SysApi
    │
    ├── SysLoginLog (无 tenant_id)
    └── SysOperationLog (无 tenant_id)
```

## 六、ID 精度处理

前端 JavaScript 的 Number 类型只能安全表示 -2^53 ~ 2^53 的整数。Go 的 int64 最大值 2^63 会导致精度丢失。

**规则：** 所有返回前端的 ID 字段在 gRPC proto 中定义为 `string` 类型，Logic 层用 `strconv.FormatInt()` / `strconv.ParseInt()` 转换。

```protobuf
message UserResponse {
    string id = 1;       // int64 → string，避免精度丢失
    string username = 2;
}
```

```go
// Logic 层
id, _ := strconv.ParseInt(in.Id, 10, 64)
user, _ := client.SysUser.Get(ctx, id)
return &apps.UserResponse{
    Id: strconv.FormatInt(user.ID, 10),
}
```

## 七、关键设计决策

| 决策 | 选项 | 选择理由 |
|---|---|---|
| 租户隔离 | 行级 tenant_id | 简单、灵活、不需要独立数据库 |
| 字典继承 | tenant_id=0 表示系统默认 | 无需额外关联表，查询逻辑简洁 |
| 审计字段 | Mixin Hook 自动填充 | 不依赖开发者手动设值，防止遗漏 |
| 软删除 vs 硬删除 | 软删除（deleted_at） | 可恢复，数据可审计 |
| 服务间通信 | gRPC | 高性能、强类型、双向流 |
| **API 鉴权** | **Casbin Domain RBAC + casbin_rule 表** | 原生多租户支持，keyMatch 通配，无需维护关联表 |
| 菜单权限 | sys_role_menus 关联表（ent edge） | Casbin 不管理前端菜单 |
| **角色-API 关联** | **Casbin 策略（casbin_rule 表）** | 不创建 sys_role_apis 表，避免双写同步 |
| 代码生成 | Ent + goctl (--style goZero) | 驼峰命名，标准化 |
| ID 返回前端 | string 类型 | 避免 int64 精度丢失 |
| JSON 字段 | camelCase | 前端 JavaScript 惯例 |
| JWT 中间件位置 | Basedata API 层 | Gateway 纯代理不做鉴权，basedata API 统一解析 |
| Casbin 执行位置 | Basedata API 层 | JWT 解析后直接检查，无需额外 RPC 调用 |
