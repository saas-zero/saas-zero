# SaaS-Zero 基础数据服务 (Basedata)

基于 go-zero 构建的多租户微服务版本 —— **核心业务与数据服务**

地址：https://github.com/saas-zero/saas-zero-basedata

## 职责

采用 **HTTP API（:18083）+ gRPC RPC（:18084）** 双层架构：

### API 层（对外 HTTP 接口）

- 暴露全部管理后台 CRUD：用户 / 角色 / 菜单 / 部门 / 字典 / 字典数据 / 租户 / 套餐 / API / 日志
- 内置三大中间件：
  - **JWT 中间件**：解析 token + Redis 会话校验，注入当前用户 / 租户上下文
  - **Casbin 中间件**：遍历角色码执行 Domain RBAC 检查，全拒返回 403
  - **操作日志中间件**：非 GET 请求自动记录操作日志（模块 / 参数 / 耗时 / IP）
- 初始化 API：`/init/all` 一键初始化（菜单 / 套餐 / 租户 / 角色 / 用户 / Casbin 策略）

### RPC 层（内部 gRPC 服务）

- **Ent 数据访问**：全部业务表 CRUD，ent 自动迁移建表
- **Casbin 策略管理**：`AssignApis` 写入 / 清理 `casbin_rule` 策略
- **自动审计**：Mixin Hook 从 context 自动填充雪花 ID、审计字段、租户 ID、软删除信息