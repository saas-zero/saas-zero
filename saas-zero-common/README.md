# SaaS-Zero 公共库 (Common)

基于 go-zero 构建的多租户微服务版本 —— **项目公共库**

地址：https://github.com/saas-zero/saas-zero-common

> 没有 `main`，仅作为库（`pkg/`）被各服务引入使用。

## 包含能力（`pkg/`）

| 包 | 功能 |
|---|---|
| `ent/mixins` | Ent 可复用混入字段：雪花 ID、租户、审计、软删除、状态、排序、备注 |
| `snowflake` | 雪花 ID 生成器（`SNOWFLAKE_WORKER_ID` 可配） |
| `bcrypt` | 密码哈希与验证 |
| `jwt` | JWT 签名与解析（含 roleCodes / tokenVersion） |
| `crypto` | AES-GCM 加解密（敏感字段脱敏） |
| `casbin` | Casbin Domain RBAC 模型 + PostgreSQL adapter |
| `errno` | 统一业务错误码 |
| `id` | int64 ↔ string 转换（前端精度无损） |
| `pagination` | 分页参数标准化 |
| `redis` | Redis 客户端封装 |
| `captcha` | 图形验证码（base64） |
| `timex` | 时间格式化工具 |