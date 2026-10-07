[English](TESTING.md) | [简体中文](TESTING.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# 测试

串行、限制并发执行。普通单元覆盖 memory 事务、严格 HTTP/JSON、名单/build、不可变重试、核对、超时/中止、local/BC 信任和历史 wire 兼容；不是浏览器/WebRTC/原生英雄技能认证。

```sh
GOMAXPROCS=2 go test -p=1 ./cmd/... ./internal/...
GOMAXPROCS=2 go vet -p=1 ./cmd/... ./internal/...
python3 scripts/test-production-v13.py
```

可选 `GOMAXPROCS=2 go test -race -p=1 ./cmd/... ./internal/...` 需平台 race 运行时。memory 测试不设置 DUEL_TEST_DSN。profile 脚本使用 GOPROXY=off 和临时 Go overlay，通过 memory 测试检查保留生产名单；本地 Go cache 须已有已验证依赖，不创建数据库/服务。

明确选择 `cmd` 和 `internal`：`deploy/production-v13/heros22_profile_test.go` 是 `duel` 包的 overlay 夹具，不是独立 Go 包。仓库级 `./...` 会包含夹具目录而编译失败；profile 脚本提供所需 overlay。

## MySQL 集成

`internal/duel/mysql_test.go` 中已有测试要求 DUEL_TEST_DSN 指向**专用可丢弃 Unix socket**，路径含 `duel-mysql-test-` 或 `.mysql-integration`，schema 只允许 `dota_duel`。测试拒绝 TCP/无关或保留 QA socket，应用迁移且每 case 清专用表。开发者需另行准备该环境，不用生产或有保留记录的库。

SQL 检查真实事务/锁/重试、view/生成字段、身份成员与维护，通过实际 driver；无显式 DSN 时可能跳过 integration，需如实说明。破坏性集成不能与同库另一 QA 并发。不公开 DSN、token、玩家/比赛 ID、原始 submissions 或逐局回执，本地证据应忽略，仅提交粗粒度可审核摘要。

## 本地只读探测与文档

开发指南中的 health/registry/OPTIONS 检查 CORS/profile 可见性，不证明自然对局。真实浏览器/设备、跨网/NAT和持久化验收需独立受控夹具与明确范围；不承诺 TURN 或反作弊。

纯文档/模板检查双语配对、语言/官网首部、相对链接、脚本路径和 `git diff --check`。固定二进制复现是构建，不是普通文档检查。当前源码删除的旧日志仍在 Git 历史，不能作当前验证。

双语维护：`python3 scripts/check-docs.py` 检查全部保留 Markdown 首部、语言配对与相对链接。
