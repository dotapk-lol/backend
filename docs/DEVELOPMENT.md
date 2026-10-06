# 本地开发与 22 英雄联调

本指南针对已有、由开发者掌控的隔离 MySQL `dota_duel`，不连接或更改生产。服务只接受该 schema 名称，运行账号仅需业务 DML；迁移/元数据管理由独立本地管理会话完成。不要把 DSN、备份或原始战绩提交进仓库。

## 迁移与名单准备

在已存在的本地开发 schema 上，按顺序应用 `migrations/001_init.sql`、`002_analytics.sql`、`003_local_pvp_analytics.sql`、`004_hero_registry.sql`。以下仅适用于尚未登记 22 名单的本地初始化状态；`--login-path=<LOCAL_ADMIN_PROFILE>` 是开发者自行配置的本地 MySQL CLI 凭据配置，占位符不是共享账号：

```sh
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < migrations/001_init.sql
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < migrations/002_analytics.sql
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < migrations/003_local_pvp_analytics.sql
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < migrations/004_hero_registry.sql
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < deploy/production-v13/register-heros22.sql
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < deploy/production-v13/append-room-fix-runtime.sql
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < deploy/production-v13/append-i18n-runtime.sql
```

这些文件已存在，本文没有新增导出/注册工具。先检查本地 schema 状态；register-heros22 是 INSERT，不应重复运行，后两项是精确前置列表的 compare-and-swap。预期 register 为 1 roster + 22 members，两个 append 各更新 1 行；0 行意味着前置状态不符，不能当作成功。已有数据库需逐项审阅缺失状态，不整批重放部署 SQL。本轮不执行这些命令，也不新增数据库。

## Go embed profile

普通 `go run ./cmd/dueld` / `go build` 使用 `internal/duel/registry/gameplay-rosters.json`，只包含 legacy20；设置环境变量不能启用新的名单。现有 `deploy/production-v13/gameplay-rosters.json` 是完整历史 production profile。Go overlay 可在不改源文件的情况下替换 embed 输入。以下从仓库根目录生成本地临时 overlay 并运行服务：

```sh
python3 - <<'PY'
import json, pathlib, tempfile
root = pathlib.Path.cwd()
p = pathlib.Path(tempfile.gettempdir()) / 'dotapk-local-roster-overlay.json'
p.write_text(json.dumps({'Replace': {
    str(root / 'internal/duel/registry/gameplay-rosters.json'):
    str(root / 'deploy/production-v13/gameplay-rosters.json')
}}))
print(p)
PY
export DUEL_MYSQL_DSN_FILE='<ABSOLUTE_PATH_TO_LOCAL_DSN_FILE>'
export DUEL_LISTEN='127.0.0.1:18082'
export DUEL_ALLOWED_ORIGIN='http://127.0.0.1:4173'
GOMAXPROCS=2 go run -p=1 -overlay="${TMPDIR:-/tmp}/dotapk-local-roster-overlay.json" ./cmd/dueld
```

overlay 路径以 Python 输出为准；若 Python tempfile 的目录与 shell TMPDIR 不同，将 `-overlay` 替换为输出的绝对路径。需要二进制时可用 `GOMAXPROCS=2 go build -p=1 -overlay='<ABSOLUTE_OVERLAY_PATH>' -o '<LOCAL_BINARY_PATH>' ./cmd/dueld`。不要覆盖正在运行的服务。该文件仅描述现有 embed 替换，不含凭据或数据，停止进程后可删除。

production profile 还保留历史 legacy20/46 绑定以读取旧记录；它不代表当前前端开放了 46 英雄。前端 profile 仍只准入 22/88，不恢复暂停英雄。SQL roster 元数据须与 embed 一致，否则持久化/分析身份可能不完整。

## build 绑定

| 字段 | 当前 22 英雄 profile |
| --- | --- |
| rosterId | `arena-heros22-v1` |
| registryVersion | `duel-heroes-127-v1` |
| registrySha256 | `5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138` |
| 允许 gameVersions | `duel-27c78aa4cfc8facc8a23`、`duel-6b1d12f75aa4bbac4e12`、`duel-851e67d77307f479f1fa` |
| 当前前端源码 NET_VERSION | `duel-851e67d77307f479f1fa` |
| heroIds | `1,3,4,5,7,8,9,15,17,18,28,31,32,36,50,55,57,58,62,71,81,82` |

`registryNumericId` 与 Valve ID、数组位置不同。GET `/api/v1/registry` 返回固定身份与 gameplayRosters；目录身份不等于 playable roster。新 roster 无通配符，当前 build 省略 rosterId 会 409，不在名单的英雄会 400，host/guest 必须同 build/roster。前端还检查完整注册表 hash 与所有身份字段，详见 [前端开发指南](https://github.com/dotapk-lol/frontend/blob/main/docs/DEVELOPMENT.md)。

重新运行前端 build 会生成新的 gameVersion；它可能不在当前 allowlist。必须在独立本地 profile 副本及本地 SQL 元数据中人工审阅登记同一个精确版本，并重新运行匹配的 backend；不要伪造旧 NET_VERSION、把 127 目录当 rosterId 或为生产加通配。文档提交不会自动更新游戏 build。Go 不执行 heros 技能，公共 88 assembly hash 与前端组合 hash 不同，详见 [heros 兼容性](https://github.com/dotapk-lol/heros/blob/main/docs/release-compatibility.md)。

## CORS 与最小检查

前端按其 README 启动 `http://127.0.0.1:4173`；必须与 DUEL_ALLOWED_ORIGIN 精确一致。localhost 与 127.0.0.1 不等价，地址末尾不要加路径或 `/`。本地直连无需 DUEL_TRUSTED_PROXY_IP。

```sh
curl --fail http://127.0.0.1:18082/healthz
curl --fail -H 'Origin: http://127.0.0.1:4173' http://127.0.0.1:18082/api/v1/registry
curl --fail -X OPTIONS -H 'Origin: http://127.0.0.1:4173' -H 'Access-Control-Request-Method: POST' -H 'Access-Control-Request-Headers: Content-Type, Authorization' http://127.0.0.1:18082/api/v1/sessions
```

这只检查本地 health/registry/预检；不生成玩家报告。进入人机或建房前确认返回名单/版本，结果由匿名会话鉴权，完整 shapes 见 [API.md](API.md)。没有 WebSocket 信令或公共 analytics REST endpoint。WebRTC 无 TURN，跨网络需要真实设备/NAT 验证；源码测试不替代这些验证。

## 测试与统计边界

README 中的 Go 单元测试不需要真实数据库；`scripts/test-production-v13.py` 仅用临时 overlay 测 production profile 准入，拒绝模式/build 不符，串行低内存运行。MySQL integration 会清表，只能用 TESTING.md 明确允许的可丢弃 socket 环境。

`duel_hero_balance_v3` 按 game_version、roster/registry、mode、status、trust、transport、AI 难度、seat、hero/opponent 分组。PVP 双 seat 可形成两个 appearance，PVE 只统计人类 seat 0；`appearances` 不等于唯一比赛数。质量 view 不自动纠正异常，汇总前仍需检查重复/时间/身份/可信度/测试局。公开平衡数据只按 [balance-data 说明](https://github.com/dotapk-lol/heros/blob/main/balance-data/README.md) 内部手动整理和人工审阅；不提交原始 SELECT * 结果、match ID 或报告 body。

## 许可范围

本项目自有代码与开发者文档采用 [MIT](../LICENSE)；修改或分发时保留版权和许可声明。第三方图片、音乐、字体、商标与依赖各自适用的许可保持独立，不包含在项目自有代码 MIT 授权内。保留上游 LICENSE/NOTICE 与来源记录；素材是否可再分发应按素材自身授权核实，不能仅凭项目 LICENSE 判断。
