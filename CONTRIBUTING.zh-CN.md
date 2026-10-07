[English](CONTRIBUTING.md) | [简体中文](CONTRIBUTING.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# 贡献指南

从最新 main 建独立分支，先读 [API](docs/API.zh-CN.md)、[开发](docs/DEVELOPMENT.zh-CN.md)、[架构](docs/ARCHITECTURE.zh-CN.md)，说明触发条件、修改后行为和实际验证。

- 中英 Markdown 同时维护，首部语言/官网链接和相对链接正确；不翻译/复制 schema、profile 数据或标准英文 MIT 法律文本。
- 保留严格 shape、匿名鉴权、幂等/不可变报告和历史 wire 身份，不把 client_reported 升为 peer_agreement。
- 保持固定 ID 分配与明确 embed/SQL build绑定。目录存在不等于可玩，未独立验收前仍22前端。
- schema/view 改动说明迁移、历史兼容、粒度/分母与异常排除，不新增数据库规避现有 MySQL 定义。
- 按范围串行测试；破坏性集成只用专用可丢弃白名单 socket，不能生产/保留 QA 或共享并发实例。
- 不提交 DSN、token、玩家/比赛标识、IP/房间码、原始报告、可追踪回执或备份；公开人工汇总按 heros/balance-data 说明。
- 自有代码/文档采用 [MIT](LICENSE)，保留声明和第三方许可，不新增未授权媒体。

检查 `git diff --check`、双语/链接及适用测试。部署/安全/DB/CI权限需明确范围；协调并发，不强推覆盖历史。文档清理无需生产操作或部署。
