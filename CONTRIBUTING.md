# 贡献指南

从最新 main 建独立分支，先阅读 [README](README.md)、[本地联调](docs/DEVELOPMENT.md) 与 [API 合约](docs/API.md)。PR 说明实际问题、修改后的行为、相关命令与验证限制。

- 保留 API 严格 JSON shapes、匿名鉴权、幂等/不可变报告与历史身份；不得把 client_reported 升为 peer_agreement。
- 固定 registry ID 不重排，不使用数组位置或 Valve ID 替代；新 roster/build 须在 embed 与 SQL 中明确登记，当前前端仍只有 22 英雄。
- schema 与统计变化需解释迁移、历史记录兼容、grain/分母和异常局排除；不新增数据库来绕过现有 MySQL 口径。
- 单元/profile 测试逐条低内存执行；数据库 integration 会清表，只用于专用可丢弃实例，不能指向生产、保留 QA 库或与其他用例并发。
- 提交前 `git diff --check`，核对文档链接和命令。文档 PR 无需生产验证或全量构建；不以历史部署证据宣称当前游戏验收。
- 不提交密码、DSN、token、playerId、IP、房间码、原始报告、细粒度追踪数据或生产备份；统计快照仅按 heros/balance-data 的人工流程汇总审阅。
- 自有代码与文档按 [MIT](LICENSE) 贡献，保留版权/许可声明；第三方依赖、图片、音乐与商标遵循各自许可，保留上游 LICENSE/NOTICE，不擅自改授 MIT 或新增未审阅媒体。

部署、DB、安全配置、CI 权限变化应明确提出独立范围，不从文档补全推导操作权限。遇到并发修改先协调，使用正常快进或 PR，不强推覆盖历史。
