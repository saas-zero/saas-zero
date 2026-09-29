# SaaS-Zero Auth
基于zero构建的多租户微服务版本  

认证服务 — OAuth 登录 / JWT 签发 / 令牌验证。
地址：https://github.com/saas-zero/saas-zero-auth  

| 属性 | 值 |
|---|---|
| 端口 | `:18081`（HTTP API） |
| 路由前缀 | `/oauth/*` |
| 入口文件 | `api/authapis.go` |
| 配置 | `api/etc/authApis.yaml` |

## 功能

| 端点 | 说明 |
|---|---|
| `POST /oauth/login` | 登录（bcrypt 验证 + JWT 签发） |
| `GET /oauth/verify` | 令牌验证（查 Redis 确保未被注销） |
| `POST /oauth/refresh` | 令牌刷新（携带 TokenVersion） |
| `GET /oauth/userinfo` | 当前用户信息（调 basedata gRPC） |
| `GET /oauth/menus` | 用户菜单树 |
| `GET /oauth/permissions` | 用户权限标识 |
| `POST /oauth/password/change` | 修改密码 |
| `POST /oauth/password/reset` | 重置他人密码 |
| `GET /oauth/code` | 图形验证码（存 Redis，TTL 300s） |

## 登录流程

```
POST /oauth/login {"tenantCode":"default","username":"admin","password":"***","captchaId":"...","captchaVal":"..."}
  │
  ├─ 1. captchaId 非空 → 从 Redis 校验验证码 → 删除
  ├─ 2. gRPC → basedata: GetTenantByCode → tenantId
  ├─ 3. gRPC → basedata: GetUserByUsername(tenantId, username)
  ├─ 3.5 锁定检查 + 状态检查（status != active → 拒绝登录）
  ├─ 4. bcrypt.Verify(password, user.Password)
  ├─ 5. 读取 token_version:{userId}（不递增，多端共存）→ Setex 确保 key 创建
  └─ 6. jwt.Sign → SETEX token:{jti}（Redis）→ 返回 token
```

> **token_version 策略**：登录**不 INCR**（避免多标签/多端互相踢出），只读取当前版本并 `Setex` 确保 key 存在（首次登录，否则 jwtauth 校验 key 不存在误判 401）。权限变更（分配角色/API、角色改 code、禁用角色、改用户角色）、密码变更、禁用/删除用户时会 INCR 使旧 token 失效。

## 配置项

```yaml
JwtSecret: saas-zero-secret-key-2024  # JWT 签名密钥
JwtExpire: 86400                       # Token 过期秒数
Redis:
  Host: 127.0.0.1:26379
  Pass: "Redis.123456"
  Type: node
  DB: 0                                # 0=go-zero, >0=go-redis
BaseDataRpc:
  Etcd:
    Hosts: ["127.0.0.1:22379"]
    Key: basedataservice.rpc
```

## 启动

```bash
# 从 workspace 根
go run ./apps/saas-zero-auth/api
# 或进入目录
cd apps/saas-zero-auth/api
go run authapis.go -f etc/authApis.yaml
```

## 依赖

- `saas-zero-common` — JWT / bcrypt / Redis 封装 / 错误码
- `saas-zero-basedata` — gRPC 调用用户/租户/菜单数据
- Redis — 验证码存储、Token 黑名单、TokenVersion
