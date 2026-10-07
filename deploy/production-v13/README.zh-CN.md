[English](README.md) | [简体中文](README.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# 固定 gameplay profile

[gameplay-rosters.json](gameplay-rosters.json) 保留明确的 legacy20、旧 `arena-first22-46-v1` 和已发 `arena-heros22-v1` 绑定；不解锁当前前端46/127，前端仍22/88。默认 source embed 仍 legacy20。

固定复现工具输入：

| 输入 | 身份 |
| --- | --- |
| Go source commit | `aa04e01f25ef2226b213cfc64e68c6c96fd0e18c` |
| Profile SHA256 | `405af9bf209f8d343e4d8e3dd8f2beaa88ec22b4fe06fe8b64e0a5785bad592b` |
| 预期 Linux 二进制 SHA256 | `94e0bbb7bd6bddefab84e15fd3ff7db53f47245813dd0a826ca50088d637f5ca` |
| 工具链 | Go1.26.1、CGO关闭、linux/amd64、已有模块cache |

根目录运行 `python3 scripts/build-production-v13.py` 会将实际源码与该提交比较，以 overlay 复现产物，不访问数据库/生产或重启服务。不一致需查原因，不改预期 digest；这是历史产物身份，不保证当前在线。

迁移004只初始化 legacy20，`register-roster.sql` 添加审核后的旧46，`register-heros22.sql` 添加22，append-runtime 文件保留审核后的精确 build列表。登记依赖匹配 schema 前置条件，INSERT 不能盲目重跑；运行 profile 与 SQL 元数据须一致，不用通配 build 或数组位置作为身份。

`heros22_profile_test.go` 由 `scripts/test-production-v13.py` 在 memory 下应用，不是游戏/浏览器测试。本地使用、运维审阅与回滚见[开发](../../docs/DEVELOPMENT.zh-CN.md)、[部署](../../docs/DEPLOYMENT.zh-CN.md)、[测试](../../docs/TESTING.zh-CN.md)。工作树不保留旧逐局/provision 回执，原 Git 历史仍在。
