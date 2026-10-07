[English](README.md) | [简体中文](README.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# DOTA DUEL Go / MySQL 后端

[dotapk.lol](https://dotapk.lol) 的 API：匿名会话、六位邀请码、WebRTC 信令、不可变报告核对与分层结果统计。玩家无需登录账号；战斗在浏览器，Go 不执行英雄技能。

| 仓库 | 职责 |
| --- | --- |
| [frontend](https://github.com/dotapk-lol/frontend) | Cloudflare 静态客户端、浏览器世界/host、UI/输入/渲染/AI 与 P2P |
| backend（本仓库） | `https://api.dotapk.lol/api/v1`、Go 服务与 MySQL 持久化 |
| [heros](https://github.com/dotapk-lol/heros) | MIT 规则/参数/host 合约；[人工汇总快照](https://github.com/dotapk-lol/heros/blob/main/balance-data/README.zh-CN.md) |

生产复用现有主机、MySQL8.4 和 Nginx，使用独立 `dota_duel` schema。API 已经使用 MySQL，不用换库、增加数据库或自动平衡导出；源码公开不授权操作生产。

## 本地启动

Go1.26.0+（go.mod）、MySQL8.4+，profile 工具可选 Python3，依赖由 go.sum 固定。已有隔离本地 schema、已应用迁移和运行 DML 账号后：

```sh
export DUEL_MYSQL_DSN_FILE='<ABSOLUTE_PATH_TO_LOCAL_DSN_FILE>'
export DUEL_LISTEN='127.0.0.1:18082'
export DUEL_ALLOWED_ORIGIN='http://127.0.0.1:4173'
go run ./cmd/dueld
```

DSN 文件格式：`<LOCAL_DB_USER>:<LOCAL_DB_PASSWORD>@tcp(127.0.0.1:3306)/dota_duel`。占位符私下填写，不提交/打印真实凭据。文件优先于 `DUEL_MYSQL_DSN`；`DUEL_LISTEN` 默认127.0.0.1:18082，`DUEL_ALLOWED_ORIGIN` 仅一个精确 origin，`DUEL_TRUSTED_PROXY_IP` 可选一个精确代理地址、默认禁用。GET `/healthz` 实际 ping MySQL，服务器没有内存 fallback。

普通构建**只含 legacy20**。当前前端为**22英雄/88槽**、`arena-heros22-v1`、`duel-heroes-127-v1`、源码 build `duel-851e67d77307f479f1fa`。22 overlay、SQL 元数据与精确绑定见[开发](docs/DEVELOPMENT.zh-CN.md)。目录身份不等于可玩，未发布英雄继续暂停/灰禁。

## 文档与检查

- [架构/目录](docs/ARCHITECTURE.zh-CN.md)
- [开发/CORS/build 登记](docs/DEVELOPMENT.zh-CN.md)
- [API 合约](docs/API.zh-CN.md)
- [测试](docs/TESTING.zh-CN.md)
- [部署与固定 profile](docs/DEPLOYMENT.zh-CN.md)、[profile 细节](deploy/production-v13/README.zh-CN.md)
- [贡献](CONTRIBUTING.zh-CN.md)

逐条执行：

```sh
GOMAXPROCS=2 go test -p=1 ./cmd/... ./internal/...
GOMAXPROCS=2 go vet -p=1 ./cmd/... ./internal/...
python3 scripts/test-production-v13.py
```

单元/profile 使用 test-only memory store，MySQL integration 需要专用、可丢弃、白名单 Unix socket，会清空其表，见测试说明，绝不能指向生产或保留 QA 数据。纯文档不用完整测试、构建或部署。

PVP/WebRTC `confirmed / peer_agreement` 是双份完成报告完全一致，不是反作弊证明；PVE/local/BC `recorded / client_reported` 分层统计，BC 仅房主上报一次。中止/争议/未完成和异常局不计胜率。内部30秒清理不是 balance-data 同步；公开快照需人工审阅汇总，不含标识/原始报告。统计 view 没有公开未认证 REST/导出接口。

自有代码/文档采用 [MIT](LICENSE)，版权2026 dotapk-lol contributors。第三方图片、音乐、商标和依赖保留各自许可/声明，不包含在此授权内，不表示 Valve 背书。
