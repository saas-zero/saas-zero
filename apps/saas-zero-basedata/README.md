# SaaS-Zero Basedata
基于 zero 构建的多租户微服务版本 — 基础数据服务  

基础数据服务 — 维护所有业务表的 CRUD，对外提供 HTTP API 和对内提供 gRPC RPC。

| 属性 | 值 |
|---|---|
| HTTP API 端口 | `:18083`（路由前缀 `/system/*`、`/init/*`） |
| gRPC RPC 端口 | `:18084`（内部服务间调用） |
| 数据库 | PostgreSQL（ent ORM 自动迁移） |
| HTTP 入口 | `api/systemapis.go` |
| RPC 入口 | `rpc/basedataService.go` |

## 两个进程

### RPC 服务（`:18084`）

gRPC 内部服务，直接操作 PostgreSQL。包含所有业务 Logic。

| 目录 | 说明 |
|---|---|
| `rpc/internal/logic/sysusers/` | 用户 CRUD + 角色分配 |
| `rpc/internal/logic/sysroles/` | 角色 CRUD + 菜单/API 分配 |
| `rpc/internal/logic/sysmenus/` | 菜单 CRUD + 树形结构 |
| `rpc/internal/logic/sysdepts/` | 部门 CRUD + 树形结构 |
| `rpc/internal/logic/sysdicts/` | 字典 CRUD |
| `rpc/internal/logic/sysdictdatas/` | 字典数据 CRUD |
| `rpc/internal/logic/systenants/` | 租户 CRUD |
| `rpc/internal/logic/syspackages/` | 套餐 CRUD |
| `rpc/internal/logic/sysapis/` | API 目录 CRUD |
| `rpc/internal/logic/syslogs/` | 日志查询 |
| `rpc/internal/logic/sysinit/` | 系统初始化（事务） |

`rpc/basedataService.go` 中通过 gRPC 拦截器从 metadata 提取用户信息注入 context：

```go
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    // 从 metadata 读取 x-user-id / x-user-name / x-tenant-id
    // → mixins.SetCurrentUserId / SetCurrentUserName / SetCurrentTenantId
}
```

### API 服务（`:18083`）

HTTP 对外接口，通过 gRPC 调用 RPC 服务。内置两层中间件：

```
请求 → JWT 中间件 → Casbin 中间件 → Logic → gRPC → RPC
  │                    │
  │ /init/* 跳过       │ /init/* 跳过
  │                    │
  └─ 解析 JWT          └─ 遍历用户 roleCodes → Enforce
     claims → context     全部拒绝 → 403
```

#### 中间件

| 中间件 | 位置 | 功能 |
|---|---|---|
| `JwtAuth` | `api/internal/middleware/jwtauth.go` | JWT 解析 + Redis token 验证 + TokenVersion 校验（始终校验） |
| `CasbinAuth` | `api/internal/middleware/casbinauth.go` | 从 JWT claims 读 roleCodes → Casbin Enforce（不额外调 gRPC） |

初始化路由（`/init/*`）跳过认证，自动注入 `userId=1, userName=system, tenantId=1`。

### 策略自动重载

API 服务启动后每 30 秒调用 `enf.LoadPolicy()` 从 `casbin_rule` 表重载策略；角色分配 API 后立即 `LoadPolicy()`。

### 操作日志中间件

`OperationLog` 中间件（JwtAuth 之后注册）异步记录所有**非 GET、非 /init/** 的写操作（新增/修改/删除），包含操作人/IP/耗时；登录操作走独立的登录日志表。

## 数据库表

11 张业务表由 **ent 自动迁移**创建（`serviceContext.go` 中 `client.Schema.Create()`）；`casbin_rule` 由 Casbin PostgreSQL adapter 自动创建（非 ent 管理）：

| 表 | Mixin | 说明 |
|---|---|---|
| `sys_users` | Base+Tenant+Created+Updated+Deleted+Status+Remark | 用户 |
| `sys_tenants` | Base+Created+Updated+Deleted+Status+Remark | 租户 |
| `sys_roles` | Base+Tenant+Created+Updated+Deleted+Status+Sort+Remark | 角色 |
| `sys_menus` | Base+Created+Updated+Deleted+Status+Remark+Sort | 菜单 |
| `sys_depts` | Base+Tenant+Created+Updated+Deleted+Status+Sort | 部门 |
| `sys_apis` | Base+Created+Updated+Deleted+Status+Remark | API 目录 |
| `sys_dicts` | Base+Tenant(Optional)+Created+Updated+Deleted+Status+Remark | 字典（继承） |
| `sys_dict_datas` | Base+Tenant(Optional)+Created+Updated+Deleted+Status+Remark | 字典数据（继承） |
| `sys_packages` | Base+Created+Updated+Deleted+Status+Sort+Remark | 套餐 |
| `sys_login_logs` | Base；显式 `tenant_id` | 登录日志（login_time，按租户查询） |
| `sys_operation_logs` | Base；显式 `tenant_id` | 操作日志（created_at，按租户查询） |
| `casbin_rule` | Casbin 管理（非 ent） | Casbin 策略 |

> 软删除表的唯一约束均为条件唯一索引（`WHERE deleted_at IS NULL`），迁移启用 `schema.WithDropIndex(true)`。

## 租户开通

管理后台「新建租户」（`CreateTenant`）自动完成：创建租户（关联套餐）→ 创建默认角色（code=admin，继承套餐菜单）→ 创建管理员用户（表单账号密码，分配默认角色）→ 回填 admin_id → Casbin 同步新租户 dom 的 API 策略（套餐 API）。

「编辑租户」可切换管理员（`GET /system/tenant/users` 从该租户用户中选择），不处理角色配置；到期时间支持 `yyyy-MM-dd`。

## 套餐分配

套餐管理支持「分配菜单 / 分配API」（`/system/package/assignMenus`、`/system/package/assignApis`），维护 `sys_package_menus`/`sys_package_apis` 模板数据；**套餐的 API 不直接写 Casbin**（Casbin 主体是角色，由新建租户的默认角色转写）。

## 生成 Ent CRUD 代码

```bash
go generate ./ent
```

## 生成 HTTP API 代码（驼峰命名）

```bash
cd api
goctl api go -api basedata_service.api -dir . -style goZero
```

## 生成 gRPC 代码（驼峰命名）

```bash
cd rpc
goctl rpc protoc basedata_service.proto --go_out=. --go-grpc_out=. --zrpc_out=. -m --style goZero
```

## 启动服务

```bash
cd rpc
go run basedataService.go -f etc/basedataService.yaml
```

## 注意事项

- ent 生成的包名为结构体全小写（`SysUser` → `sysuser`），不可配置
- goctl 使用 `--style goZero` 确保文件名和方法名驼峰
- 返回前端的 ID 字段使用 string 类型，避免 int64 精度丢失
- 数据库列名 snake_case，JSON 字段 camelCase
```

## 依赖

- `saas-zero-common` — Mixin / Casbin / Redis / 错误码
- PostgreSQL — 业务数据和 Casbin 策略
- Redis — Token 验证 / TokenVersion 校验
- etcd — RPC 服务发现
