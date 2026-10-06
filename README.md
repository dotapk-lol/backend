# DOTA DUEL Go / MySQL backend

[dotapk.lol](https://dotapk.lol) 的 Go API：匿名会话、六位房间码、WebRTC 信令、不可变结果核对与分层战绩统计。玩家无需注册账号；客户端获得匿名会话 token，token 不应提交到仓库。战斗在浏览器运行，后端不模拟英雄技能。

| 仓库 | 职责 |
| --- | --- |
| [frontend](https://github.com/dotapk-lol/frontend) | Cloudflare 静态界面、浏览器战斗、输入/渲染/AI、P2P 与英雄 host |
| **backend（本仓库）** | `https://api.dotapk.lol/api/v1`，Go 服务、MySQL 持久化、房间与结果契约 |
| [heros](https://github.com/dotapk-lol/heros) | MIT 英雄规则、参数与 host 合约；[balance-data](https://github.com/dotapk-lol/heros/tree/main/balance-data) 只说明内部手动整理汇总快照的流程 |

生产运行在 AgentSquared 现有主机，复用 MySQL 8.4 与 Nginx，使用独立 `dota_duel` schema。生产部署资料是当时验收记录，不授权贡献者操作现有服务。当前 API 已采用 MySQL，不需要换数据库或增加新的统计库。本轮没有连接、导出或修改数据库。

## 环境与本地启动

需要 Go **1.26.0+**（见 go.mod）、MySQL **8.4+**；可选 Python 3 用于已有 profile 检查脚本与本地 overlay 文件。运行服务必须有 MySQL，没有内存运行 fallback；单元测试有 test-only memory store。mysql driver 由 go.sum 固定。

已有隔离本地 `dota_duel`、迁移及最小权限运行账号准备好后：

```sh
export DUEL_MYSQL_DSN_FILE='<ABSOLUTE_PATH_TO_LOCAL_DSN_FILE>'
export DUEL_LISTEN='127.0.0.1:18082'
export DUEL_ALLOWED_ORIGIN='http://127.0.0.1:4173'
go run ./cmd/dueld
```

DSN 文件内容格式为 `<LOCAL_DB_USER>:<LOCAL_DB_PASSWORD>@tcp(127.0.0.1:3306)/dota_duel`，占位符须由本地操作者填写；不要在仓库内保存真实值或在日志中打印。`DUEL_MYSQL_DSN_FILE` 优先于 `DUEL_MYSQL_DSN`；后者仅适合临时本地环境。以上普通构建只启用 **legacy20**；要接当前 22 英雄前端，按 [DEVELOPMENT.md](docs/DEVELOPMENT.md) 使用现有 production profile overlay、SQL 元数据与精确 build 绑定。

| 环境变量 | 作用 |
| --- | --- |
| `DUEL_MYSQL_DSN_FILE` | 专用 DSN 文件，推荐；schema 必须是 dota_duel |
| `DUEL_MYSQL_DSN` | 未设置文件时使用的 DSN |
| `DUEL_LISTEN` | 默认 `127.0.0.1:18082` |
| `DUEL_ALLOWED_ORIGIN` | 单个精确前端 origin；localhost、127.0.0.1、端口/协议不等价 |
| `DUEL_TRUSTED_PROXY_IP` | 一个确切代理地址，默认禁用；本地直连不要设置 |

`GET /healthz` 实际 Ping MySQL。CORS 不允许通配/多 origin；请求带 Origin 时必须精确匹配，预检允许 GET/POST/DELETE/OPTIONS 与 Content-Type/Authorization。详细接口和字段见 [API.md](docs/API.md)，没有公开未认证的统计查询或数据导出接口。

## 当前身份与信任口径

当前前端仅发布 **22 英雄 / 88 槽**，使用 `arena-heros22-v1`、`duel-heroes-127-v1`；当前源码 build 为 `duel-851e67d77307f479f1fa`。后端 default embed、production profile、SQL 元数据是不同层，不能只改一个文件宣称已完成接线。精确成员、历史 allowlist 和本地联调步骤见 [开发指南](docs/DEVELOPMENT.md)。127 个目录 ID 不代表全部可玩，其余英雄保持灰禁、暂停适配。

- WebRTC PVP 双方结果一致：`confirmed / peer_agreement`，不是反作弊证明；单报未获双方确认不能混入该组。
- PVE、同屏 PVP、BroadcastChannel PVP：`recorded / client_reported`，分 mode、transport、trust、build、roster、英雄/对手分析；BC 仅房主上报。
- `aborted`、`disputed`、未完成或异常局不计胜率。维护任务约每 30 秒回收过期信令/会话等，这是服务内部清理，**不是**向 heros 同步数据。

迁移已有 v2/v3 平衡与质量 views。内部人员可用获授权的只读 SQL 会话进行统计；任何公开快照必须汇总并人工审阅，不能公开 playerId、token、IP、房间码、原始报告或可追踪细粒度记录。当前不提供自动同步、导出脚本或同步频率承诺。

## 目录与测试

| 路径 | 内容 |
| --- | --- |
| `cmd/dueld/` | 环境读取、HTTP 启动、MySQL 检查与清理循环 |
| `internal/duel/` | service、HTTP、store、结果 reconciliation、稳定身份/roster 与测试 |
| `migrations/001_init.sql` 至 `004_hero_registry.sql` | 持久化、统计与身份表/views，由操作者 out-of-band 应用 |
| `deploy/production-v13/` | 已审核 production profile、22 名单与历史 build 注册 SQL |
| `deploy/` | 历史 Nginx/systemd 部署资料；不是通用开发默认配置 |
| `scripts/`、`docs/` | 轻量检查、历史隔离 QA/部署证据及 API 合约 |
| `Dockerfile`、`compose.yaml` | 可选隔离本地/测试栈；不用于现有 AgentSquared 部署 |

低内存环境逐条运行：

```sh
GOMAXPROCS=2 go test -p=1 ./...
GOMAXPROCS=2 go vet -p=1 ./...
python3 scripts/test-production-v13.py
```

不要设置 DUEL_TEST_DSN 就可运行 memory 单元测试；生产 profile 脚本使用 in-memory 测试、临时 overlay，不访问 MySQL。需要 race 检查时用 `GOMAXPROCS=2 go test -race -p=1 ./...`。真实 MySQL integration 只接受专用可丢弃 Unix socket，会应用迁移并清表；严格按 [TESTING.md](docs/TESTING.md)，绝不能用生产或保留的 QA 库。文档修改只需链接/命令存在性和 `git diff --check`，不用创建数据库、浏览器或全量构建。

查看 [贡献指南](CONTRIBUTING.md)、[API](docs/API.md)、[历史部署记录](docs/DEPLOYMENT.md) 与 [验证记录](docs/TESTING.md)。

## 许可

本项目自有代码与开发者文档采用 [MIT 许可](LICENSE)，版权归属为 `Copyright (c) 2026 dotapk-lol contributors`。使用、修改和分发时保留许可及版权声明。第三方依赖遵循各自许可，保留其 LICENSE/NOTICE；项目 MIT 不取代依赖许可。Valve 名称、商标、图像、音乐及其他第三方素材不纳入自有代码 MIT，复用时分别核对授权，不表示 Valve 背书。[heros](https://github.com/dotapk-lol/heros/blob/main/LICENSE) 保持其独立 MIT 许可。
