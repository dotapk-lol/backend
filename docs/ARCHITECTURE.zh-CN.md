[English](ARCHITECTURE.md) | [简体中文](ARCHITECTURE.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# 架构与数据粒度

HTTP handler 验证匿名会话、严格 body/origin/成员身份并实施持久化请求限流。service 事务管理房间座位、幂等、准备、超时和不可变报告核对；SQLStore 持久化到现有 MySQL，只有测试使用 memory store。浏览器战斗不属于服务端。

| 路径 | 职责 |
| --- | --- |
| `cmd/dueld/` | 环境、SQL ping、HTTP listener 与30秒维护 |
| `internal/duel/http.go`、`service.go`、`local.go` | 路由、会话/房间/比赛生命周期与 local 上报 |
| `internal/duel/reconcile.go`、`model.go` | 报告校验、摘要核对和 wire shapes |
| `internal/duel/store.go` | MySQL 事务持久化与清理 |
| `internal/duel/registry/`、`registry.go` | 冻结身份、embed 名单/build |
| `migrations/001_init.sql`–`004_hero_registry.sql` | 持久化、v1/v2统计、roster/v3元数据 |
| `deploy/production-v13/` | 保留 profile、审核后的增量登记 SQL 和 profile 测试 |
| `deploy/` | 已有操作者 Nginx/systemd/DSN隔离配置，不是通用凭据 |
| `scripts/build-production-v13.py`、`scripts/test-production-v13.py` | 固定二进制复现与 memory profile 检查 |
| `Dockerfile`、`compose.yaml` | 可选隔离本地/staging栈，不用于现有 现有主机 部署 |

记录保存 game version/roster/registry 快照。local/BC 的比赛内部座位 ID 不代表两个已鉴权的人，只由一个报告者拥有记录。PVP 每座位报告不可变，HTTP 隐藏原始 submissions，双报只有在语义核对边界一致才确认。维护过期邀请码/SDP/会话/限流及超时比赛，不导出数据。

view 除 build、身份和对手外，还按 status/trust/transport 分组。PVP 一局可有两个 seat appearance，PVE 只统计人类 seat0。原始表/view 都属内部，匿名 ID 仍敏感；汇总前查缺失映射、时间/赢家/确认异常、重复和测试局。没有自动纠错或公共统计接口，中止/争议不计胜场。

精确输入输出见 [API](API.zh-CN.md)，公开快照见[人工说明](https://github.com/dotapk-lol/heros/blob/main/balance-data/README.zh-CN.md)。删除的历史回执/日志仍在 Git 历史；源码清理不删除服务端数据或改写历史。
