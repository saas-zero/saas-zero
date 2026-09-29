# SaaS-Zero Etcd 调试工具

服务注册中心。开发环境通过 etcd 管理服务发现与配置分发。

## 启动

```bash
# 启动 etcd 服务
cd apps/saas-zero-etcd
go run etcd.go
```

## 调试

```bash
# 查看已注册的服务
etcdctl get --prefix ""

# 查看 basedata RPC 注册信息
etcdctl get /basedataservice.rpc --prefix
```

## 连接信息

| 项目 | 值 |
|---|---|
| 版本 | etcd release-3.6.5 |
| 源码 | https://github.com/etcd-io/etcd |
| 客户端端口 | 宿主 `127.0.0.1:22379`（docker 映射 2379） |

> ⚠️ **advertise 地址注意事项**：docker 运行 etcd 时 `--advertise-client-urls` 必须配置为**宿主可达地址**（如 `http://127.0.0.1:22379`），不能是容器内 `0.0.0.0:2379`。
> 否则 go-zero 的 etcd client 每 1 分钟 `AutoSyncInterval` 会把 endpoints 覆盖为容器内地址，一旦连接断开（docker 抖动等）就无法恢复，日志出现 `dial tcp 0.0.0.0:2379: connect: connection refused`，需重启各服务才能恢复。
