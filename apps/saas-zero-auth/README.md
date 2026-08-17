# SaaS-Zero 认证服务 (Auth)

基于 go-zero 构建的多租户微服务版本 —— **认证与授权服务**

地址：https://github.com/saas-zero/saas-zero-auth

## 职责

- **登录认证**：`POST /oauth/login`，支持租户编码 + 账号密码 + 图形验证码
  - bcrypt 校验密码，连续失败锁定（5 次 / 30 分钟）
  - 成功后签发 JWT（含 `userId / tenantId / userName / roleCodes / tokenVersion`）
- **令牌管理**：`/oauth/verify` 验证、`/oauth/refresh` 刷新、Redis 会话控制（踢旧会话）
- **用户上下文**：`/oauth/userinfo` 当前用户、`/oauth/menus` 用户菜单树、`/oauth/permissions` 权限码
- **账号安全**：`/oauth/password/change` 修改密码、`/oauth/password/reset` 重置密码（清除锁定）
- **验证码**：`/oauth/code` 输出 base64 图形验证码
- 通过 **gRPC** 调用基础数据服务查询租户 / 用户 / 角色码