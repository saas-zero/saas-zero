# SaaS-Zero Common
基于zero构建的多租户微服务版本  

公共库，无 main 入口，作为 Go module 被各微服务 import 使用。
地址：https://github.com/saas-zero/saas-zero-common  

## 包目录

```
pkg/
├── ent/mixins/        # Ent Schema 可复用混入字段
│   ├── base.go        # 雪花 ID（BaseMixin）
│   ├── created.go     # 创建审计（CreatedMixin）
│   ├── updated.go     # 更新审计（UpdatedMixin）
│   ├── deleted.go     # 软删除审计（DeletedMixin）
│   ├── tenant.go      # 租户隔离（TenantMixin）
│   ├── status.go      # 状态枚举（StatusMixin）
│   ├── sort.go        # 排序号（SortMixin）
│   └── remark.go      # 备注（RemarkMixin）
├── snowflake/         # 雪花 ID 生成器
├── bcrypt/            # 密码哈希与验证
├── jwt/               # JWT 签名与解析（含 TokenVersion）
├── crypto/            # AES-GCM 加解密
├── casbin/            # Casbin Domain RBAC + PostgreSQL adapter
├── captcha/           # 图形验证码生成
├── id/                # int64/string ID 安全互转（JS 精度）
├── pagination/        # 分页参数规范化（page/size/offset）
├── timex/             # Unix 时间戳格式化（秒/毫秒自动判断）
├── envconf/           # 环境变量配置加载
├── redis/             # Redis 客户端封装（DB=0 用 go-zero，DB>0 用 go-redis）
└── errno/             # 统一错误码
```

## Mixin 自动审计

| Mixin | 触发时机 | 自动填充字段 |
|---|---|---|
| `BaseMixin` | `OpCreate` | `id`（雪花 ID） |
| `CreatedMixin` | `OpCreate` | `created_at`, `created_id`, `created_by` |
| `UpdatedMixin` | `Create`/`Update`/`UpdateOne` | `updated_at`, `updated_id`, `updated_by` |
| `DeletedMixin` | `Update`/`UpdateOne`（设了 `deleted_at`） | `deleted_at`, `deleted_id`, `deleted_by` |
| `TenantMixin` | `OpCreate` | `tenant_id`（从 context 读取） |
| `StatusMixin` | — | `status` 字段（枚举约束） |
| `SortMixin` | — | `sort` 排序号字段 |
| `RemarkMixin` | — | `remark` 备注字段 |

使用前需通过 context 注入用户信息：

```go
ctx = mixins.SetCurrentUserId(ctx, 1001)
ctx = mixins.SetCurrentUserName(ctx, "admin")
ctx = mixins.SetCurrentTenantId(ctx, 1001)
```

## Redis 客户端

`redis.Conf` 通过 `DB` 字段选择实现：

- **DB=0**（默认）：走 go-zero `redis.NewRedis`，带 breaker/hooks
- **DB>0**：走 go-redis 直连

```go
import "github.com/saas-zero/saas-zero-common/pkg/redis"

client, _ := redis.NewClient(redis.Conf{
    Host: "127.0.0.1:6379",
    Pass: "",
    DB:   1,  // 非 0 则用 go-redis
})
client.Setex("key", "value", 3600)
```

## 测试

```bash
go test ./pkg/snowflake   # 雪花 ID
go test ./pkg/bcrypt      # 密码哈希
go test ./pkg/jwt         # JWT 签名
go test ./pkg/crypto      # AES-GCM
```

## 依赖

- `entgo.io/ent` v0.14.5 — Mixin 定义
- `github.com/golang-jwt/jwt/v5` — JWT
- `github.com/casbin/casbin/v2` — Casbin adapter
- `github.com/zeromicro/go-zero` v1.9.2 — Redis 封装
- `github.com/redis/go-redis/v9` — 多 DB 支持
- `github.com/mojocn/base64Captcha` — 验证码
