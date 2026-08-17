# SaaS-Zero 网关服务 (Gateway)

基于 go-zero 构建的多租户微服务版本 —— **API 网关，统一流量入口**

地址：https://github.com/saas-zero/saas-zero-gateway

## 职责

- **统一入口**：对外唯一暴露节点，监听 `:18080`
- **HTTP 纯透传代理**：按路径前缀将请求转发到对应上游服务（go-zero `gateway`）
  - `/oauth/*` → 认证服务（`:18081`）
  - `/system/*`、`/init/*` → 基础数据 API（`:18083`）
- **不做鉴权**：不解析 JWT、不做 Casbin 检查，只做路由转发，鉴权逻辑收敛在各业务服务内部