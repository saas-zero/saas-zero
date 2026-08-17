# SaaS-Zero 项目文档指南（本仓库）

> 本仓库（`github.com/saas-zero/saas-zero`）是 **SaaS-Zero 开源项目的展示仓库**——仅承载整体架构与功能介绍，**不含业务代码**。各微服务代码分别在独立仓库中维护，见下方「代码仓库索引」。

本文件约定如何维护本仓库的文档，供协作者与 AI 代理阅读。

## 仓库定位

- **给谁看：** 潜在用户、技术评估者、开源社区
- **做什么：** 用简洁清晰的方式讲清楚 SaaS-Zero「是什么、解决什么问题、怎么跑起来」
- **不做什么：** 不含源代码、不含内部部署信息、不含敏感凭据

## 目录结构

```
.
├── README.md            # 入口文档（推广主页）：特性、架构全景、技术栈、快速开始、模块链接
├── ARCHITECTURE.md      # 深度架构设计：分层、多租户、认证授权、Casbin 模型、实体关系、关键决策
├── AGENTS.md            # 本文件：文档维护约定
├── apps/
│   ├── saas-zero-gateway/README.md   # 网关模块说明（独立仓库入口）
│   ├── saas-zero-auth/README.md      # 认证模块说明（独立仓库入口）
│   ├── saas-zero-basedata/README.md  # 基础数据模块说明（独立仓库入口）
│   └── saas-zero-etcd/README.md      # Etcd 调试工具说明（独立仓库入口）
├── saas-zero-common/README.md        # 公共库说明（独立仓库入口）
├── doc/images/          # 截图 / 架构图
├── LICENSE / NOTICE     # Apache-2.0 协议与声明
└── go.work              # Go workspace（配合各模块仓库聚合构建用）
```

## 代码仓库索引

各模块独立建仓，README 中应提供对应链接：

| 模块 | 仓库 |
|---|---|
| 项目（本仓库） | https://github.com/saas-zero/saas-zero |
| 网关 | https://github.com/saas-zero/saas-zero-gateway |
| 认证 | https://github.com/saas-zero/saas-zero-auth |
| 基础数据 | https://github.com/saas-zero/saas-zero-basedata |
| Etcd 工具 | https://github.com/saas-zero/saas-zero-etcd |
| 公共库 | https://github.com/saas-zero/saas-zero-common |
| 前端 | saas-zero-web（README 中引用） |

## 关键事实基线（文档联动更新时务必保持一致）

以下事实若代码仓库变更，应同步到本仓库文档：

- 端口约定：gateway `:18080`，auth `:18081`，basedata API `:18083`，basedata RPC `:18084`
- 接口规模：9 认证 + 60 业务 + 5 初始化 = **74 个端点**，全部经 gateway 访问
- update / delete 一律 **POST**（delete 传 `{ids: [...]}`）
- 架构：go-zero Gateway 纯代理 → auth / basedata API（JWT + Casbin + 操作日志中间件）→ basedata RPC（ent DB）→ PostgreSQL
- 权限：菜单走 `sys_role_menus`（前端导航），API 走 Casbin Domain RBAC（`casbin_rule` 表）；**继承式授权**（只能授出自己拥有的权限）
- 会话：JWT + Redis `tokenVersion`（改密/重配权限后旧 token 立即失效）
- ID：雪花 ID，返回前端用 string 双字段（`id` + `idStr`）
- 数据库：11 业务表 + 4 M:N 关联表 + 1 Casbin 策略表，ent 自动迁移建表
- 数据隔离：行级 `tenant_id`；字典继承模式 `tenant_id=0` 表示系统默认
- 初始化：全新环境 `POST /init/all` 一键初始化（幂等可重跑）

## 文档编写规范

1. **面向读者**：以「读懂架构、上手启动」为目标，技术细节追求准确但不堆砌代码
2. **写法**：先讲为什么（设计动机），再讲怎么做（关键机制与调用链）
3. **敏感信息**：严禁出现内网 IP、Windows 绝对路径、真实数据库密码 / 连接串、真实密钥
4. **端口用示例 IP**：示例一律使用 `127.0.0.1` / `localhost`
5. **图表**：架构 ASCII 图保持与代码仓库一致；截图放在 `doc/images/` 并以相对路径引用
6. **一致性**：改代码仓库前，先检查本文档「关键事实基线」是否同步更新

## 模块 README 模板

各 `apps/*/README.md` 采用统一三段式（中文，简短）：

```markdown
# <模块名>

基于 go-zero 构建的多租户微服务版本 —— <一句话职责>

地址：https://github.com/saas-zero/<repo>

## 职责
- <核心功能点>
- <对外端口 / 协议>
```