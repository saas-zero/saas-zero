# SaaS-Zero

基于 go-zero + ent + Casbin 的多租户 SaaS 微服务后端框架。

主项目

所有微服务项目，按照当前目录结构去构建会清晰一些

gateway端口：18080  
auth.api端口：18081  
auth.rpc端口：18082 (暂无)  
basedata.api：18083  
basedata.rpc端口：18084

## 技术栈

| 技术 | 用途 | 版本 |
|---|---|---|
| Go | 编程语言 | 1.25.3 |
| go-zero | 微服务框架 | v1.9.2 |
| ent | ORM / 实体框架 | v0.14.5 |
| PostgreSQL | 数据库 | via lib/pq |
| gRPC | 服务间通信 | google.golang.org/grpc |
| Protobuf | 序列化 | v3 |
| etcd | 服务发现 / 配置分发 | v3.5.15 |
| JWT | 认证令牌 | golang-jwt/jwt/v5 |
| Casbin | 运行时 API 权限控制 | v2.135.0 (Domain RBAC) |

## 项目结构

```
saas-zero/
├── apps/
│   ├── saas-zero-gateway/     # API 网关 (:18080) - HTTP 纯代理
│   ├── saas-zero-auth/        # 认证服务 (:18081) - 登录 / JWT 签发
│   ├── saas-zero-basedata/    # 基础数据服务
│   │   ├── api/               # HTTP 对外接口 (:18083)
│   │   │   ├── internal/
│   │   │   │   ├── handler/   # HTTP 处理器 (goctl 生成)
│   │   │   │   ├── logic/     # 业务逻辑 (调 gRPC)
│   │   │   │   ├── middleware/ # JWT + Casbin 中间件
│   │   │   │   ├── svc/       # 服务上下文 (gRPC client + Casbin)
│   │   │   │   ├── types/     # 类型定义 (goctl 生成)
│   │   │   │   └── config/    # 配置
│   │   │   └── etc/           # YAML 配置
│   │   └── rpc/               # gRPC 内部服务 (:18084)
│   │       ├── apps/          # Protobuf 定义 (生成)
│   │       ├── client/        # gRPC 客户端包装
│   │       └── internal/
│   │           ├── logic/     # 业务逻辑 (操作 DB)
│   │           ├── server/    # gRPC 服务注册 (生成)
│   │           ├── svc/       # 服务上下文 (ent client + Casbin)
│   │           └── config/    # 配置
│   └── saas-zero-etcd/        # Etcd 调试工具
├── saas-zero-common/          # 公共库
│   └── pkg/
│       ├── ent/mixins/        # Ent 可复用混入字段
│       ├── snowflake/         # 雪花 ID 生成器
│       ├── bcrypt/            # 密码哈希与验证
│       ├── jwt/               # JWT 签名与解析
│       ├── crypto/            # AES-GCM 加解密
│       └── casbin/            # Casbin Domain RBAC + PostgreSQL adapter
├── AGENTS.md                  # AI 辅助开发指南
├── ARCHITECTURE.md            # 详细架构设计文档
└── go.work                    # Go Workspace
```

## 环境要求

- Go 1.25+
- PostgreSQL 15+
- etcd v3.5+（服务发现）

## 快速启动

### 1. 启动基础设施

确保 etcd 和 PostgreSQL 已启动。项目配置的默认连接：

```yaml
# etcd
192.168.201.52:32384

# PostgreSQL (位于 apps/saas-zero-basedata/rpc/etc/basedataService.yaml)
host=192.168.201.188 port=5432 user=postgres password=AM38xymTdFree4Fh dbname=saas_zero_kun sslmode=disable
```

### 2. 启动服务（按顺序）

```bash
# 从 workspace 根路径执行
cd D:\Projects\saas-zero\saas-zero

# 1. 基础数据 RPC 服务
go run ./apps/saas-zero-basedata/rpc

# 2. 基础数据 API 服务（等 RPC 就绪后）
go run ./apps/saas-zero-basedata/api

# 3. 认证服务
go run ./apps/saas-zero-auth/api

# 4. 网关（统一入口 :18080）
go run ./apps/saas-zero-gateway
```

### 3. 验证服务

```bash
# 健康检查
curl http://localhost:18080/oauth/code
# 预期: {"code":0,"msg":"success"}
```

## 数据库

### 业务表（11 张）

数据库表由 **ent 自动迁移** 创建。服务启动时 `client.Schema.Create()` 自动执行 DDL，无需手动建表。

| 表 | Mixin 组合 | 说明 |
|---|---|---|
| `sys_users` | Base + Tenant + Created + Updated + Deleted + Status + Remark | 用户 |
| `sys_tenants` | Base + Created + Updated + Deleted + Status + Remark | 租户 |
| `sys_roles` | Base + Tenant + Created + Updated + Deleted + Status + Sort + Remark | 角色 |
| `sys_menus` | Base + Created + Updated + Deleted + Status + Remark + Sort | 菜单 |
| `sys_depts` | Base + Tenant + Created + Updated + Deleted + Status + Sort | 部门 |
| `sys_apis` | Base + Created + Updated + Deleted + Status + Remark | API 目录 |
| `sys_dicts` | Base + Tenant(Optional) + Created + Updated + Deleted + Status + Remark | 字典 |
| `sys_dict_datas` | Base + Tenant(Optional) + Created + Updated + Deleted + Status + Remark | 字典数据 |
| `sys_packages` | Base + Created + Updated + Deleted + Status + Sort + Remark | 套餐 |
| `sys_login_logs` | Base | 登录日志 |
| `sys_operation_logs` | Base | 操作日志 |

### Casbin 策略表（1 张）

`casbin_rule` 表由 **Casbin PostgreSQL adapter 自动创建**（`CREATE TABLE IF NOT EXISTS`），无需手动建表：

```sql
CREATE TABLE casbin_rule (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL DEFAULT '',
    v0 VARCHAR(100) NOT NULL DEFAULT '',  -- roleCode
    v1 VARCHAR(100) NOT NULL DEFAULT '',  -- tenantId (dom)
    v2 VARCHAR(100) NOT NULL DEFAULT '',  -- API path
    v3 VARCHAR(100) NOT NULL DEFAULT '',  -- HTTP method
    v4 VARCHAR(100) NOT NULL DEFAULT '',  -- API ID (extra)
    v5 VARCHAR(100) NOT NULL DEFAULT ''
);
```

### 新增业务表

```bash
# 1. 在 ent/schema/ 下新建文件（参考已有表定义）
# 2. 生成 ent 代码
cd apps/saas-zero-basedata
go generate ./ent

# 3. 定义 Protobuf（仅 gRPC）
cd rpc
protoc --go_out=. --go-grpc_out=. basedata_service.proto

# 4. 或定义 HTTP API
cd ../api
goctl api go -api xxx_service.api -dir . -style goZero

# 5. 实现 Logic 层
# 重启服务后，ent 自动迁移创建新表
```

> **重要：** 使用 `protoc`（而非 `goctl rpc protoc`）生成 proto 代码，因为 goctl 的全量覆盖模式会删除手写的 Logic 代码。

## 数据初始化

系统首次启动时，数据库为空。通过初始化 API 创建种子数据：

### 初始化顺序

```
1. 创建套餐    → POST /init/package/create
2. 创建租户    → POST /init/tenant/create
3. 创建角色    → POST /init/role/create
4. 创建用户    → POST /init/user/create
```

### 示例

```bash
# 1. 创建套餐
curl -X POST http://localhost:18080/init/package/create \
  -H 'Content-Type: application/json' \
  -d '{"name":"基础版","code":"basic","status":"active","sort":1}'

# 2. 创建租户
curl -X POST http://localhost:18080/init/tenant/create \
  -H 'Content-Type: application/json' \
  -d '{"name":"测试租户","code":"test","adminId":1,"packageId":1,"status":"active"}'

# 3. 创建角色
curl -X POST http://localhost:18080/init/role/create \
  -H 'Content-Type: application/json' \
  -d '{"name":"管理员","code":"admin","status":"active","sort":1}'

# 4. 创建用户
curl -X POST http://localhost:18080/init/user/create \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123","nickname":"管理员","status":"active","roleIds":[1]}'
```

> 初始化 API 跳过 JWT 认证和 Casbin 权限检查，自动注入 `userId=1, userName=system, tenantId=1`。

## 接口列表

完整 50+ 个 API 端点，均通过 Gateway `:18080` 统一访问。

### 认证接口

| 方法 | 路径 | 说明 | 请求体 |
|---|---|---|---|
| POST | `/oauth/login` | 登录 | `{"username":"...","password":"..."}` |
| GET | `/oauth/verify` | 验证令牌 | Header: `Authorization: Bearer <token>` |
| POST | `/oauth/refresh` | 刷新令牌 | `{"token":"..."}` |
| GET | `/oauth/userinfo` | 当前用户信息 | Header |
| GET | `/oauth/menus` | 用户菜单树 | Header |
| GET | `/oauth/permissions` | 用户权限标识 | Header |
| POST | `/oauth/password/change` | 修改密码 | `{"oldPassword":"...","newPassword":"..."}` + Header |
| POST | `/oauth/password/reset` | 重置他人密码 | `{"userId":"...","newPassword":"..."}` + Header |
| GET | `/oauth/code` | 验证码 (占位) | — |

### 业务 CRUD 接口

| 资源 | 方法 | 路径 |
|---|---|---|
| 用户 | POST/PUT/DELETE/GET | `/system/user/{create,update,delete,list,detail,resetPassword,assignRoles}` |
| 角色 | POST/PUT/DELETE/GET | `/system/role/{create,update,delete,list,detail,assignMenus,assignApis}` |
| 菜单 | POST/PUT/DELETE/GET | `/system/menu/{create,update,delete,list,detail,tree,routers}` |
| 部门 | POST/PUT/DELETE/GET | `/system/dept/{create,update,delete,list,detail,tree}` |
| 字典 | POST/PUT/DELETE/GET | `/system/dict/{create,update,delete,list,detail}` |
| 字典数据 | POST/PUT/DELETE/GET | `/system/dictData/{create,update,delete,list,detail,byDictKey}` |
| 租户 | POST/PUT/DELETE/GET | `/system/tenant/{create,update,delete,list,detail,changeStatus}` |
| 套餐 | POST/PUT/DELETE/GET | `/system/package/{create,update,delete,list,detail}` |
| API | POST/PUT/DELETE/GET | `/system/api/{create,update,delete,list,detail}` |

## Casbin 权限管理

角色-API 权限通过 `POST /system/role/assignApis` 接口管理：

```bash
# 为角色分配 API 权限
curl -X POST http://localhost:18080/system/role/assignApis \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <jwt_token>' \
  -d '{"code":"admin","apiIds":[1,2,3]}'
```

该 RPC 会：
1. 清除该角色+租户的旧策略
2. 查 `sys_apis` 表获取路径和方法
3. 写入 Casbin 策略（`casbin_rule` 表）

运行时每个请求由基于 data API 的 Casbin 中间件拦截校验，任一角色通过则放行。

## 开发指南

### 编译

```bash
# 从 workspace 根编译所有模块
go build ./saas-zero-common/...
go build ./apps/saas-zero-basedata/...
go build ./apps/saas-zero-auth/...
go build ./apps/saas-zero-gateway/...
```

### 代码生成

```bash
# 生成 ent 代码（修改 schema 后执行）
cd apps/saas-zero-basedata && go generate ./ent

# 生成 gRPC 代码（修改 proto 后执行，仅 message + stub，不覆盖 logic）
cd apps/saas-zero-basedata/rpc
protoc --go_out=. --go-grpc_out=. basedata_service.proto

# 生成 HTTP API 代码（修改 .api 文件后执行）
cd apps/saas-zero-basedata/api
goctl api go -api xxx_service.api -dir . -style goZero
```

### 添加新 API 端点

1. 如果新的资源需要新表：在 `ent/schema/` 新建 → `go generate ./ent`
2. 如果新 API 需要 protobuf：在 `basedata_service.proto` 加 message + rpc → `protoc`
3. 如果新 HTTP 接口：在 `.api` 文件加路由 → `goctl api go`，或在 `routes.go` 手动加
4. 实现 logic 层
5. 如果新接口需要权限：在 `sys_apis` 表添加记录 → 通过 `assignApis` 分配给角色

## 更多文档

| 文档 | 内容 |
|---|---|
| `AGENTS.md` | AI 辅助编程指南、Casbin 用法、Mixin 说明、代码风格 |
| `ARCHITECTURE.md` | 详细架构设计、多租户、认证授权流程、实体关系 |
