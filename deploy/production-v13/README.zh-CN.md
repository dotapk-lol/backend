[English](README.md) | [简体中文](README.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# 固定 gameplay profile

## 独立28发布profile

新增六英雄（0剑圣、2帕吉、6PA、10撼地者、14风行、19潮汐）使用独立 `arena-heros28-v1`。不能扩写旧22：v3视图通过原roster解释历史成员。保留旧22的7个build、旧46的2个build、legacy20、127身份、迁移001–004和鉴权。精确28名单：`0,1,2,3,4,5,6,7,8,9,10,14,15,17,18,19,28,31,32,36,50,55,57,58,62,71,81,82`。

最终runtime `duel-6750ae2e1cfd1c3cdd42` 对应frontend `57a69c151ff388d1205a03c7e9912d375bd46ec1`、heros `7a555cdd47dc7c5eec4f0cdccf022c4a0fdbd7c0`、rulesHash `1c834e77da8f10a46a281da6df9f62458b701a7297befc9c5d5c9c52210d3f67`。复审修复包括肢解控制时长、幻影突袭反弹、六英雄效果快照来源校验和训练复活时钟；前端RULESET_HASH为 `1fb62f5882823ce2407f7f6041484285cf281b068216c93e98e41370fab7e289`。两manifest、SQL和测试只绑定此最终28版本，被替代且未发布的候选不登记生产。rulesHash核验前端工件，不存SQL。前端发布仍需单独通过真实PVP验收门槛。

`register-heros28.sql` 在一个事务中新增1行metadata和28行membership。先核对schema/registry SHA、已有legacy/46/22精确metadata及成员、28尚不存在，并备份旧二进制/profile。只执行缺失增量SQL，遇错停止（绝不用mysql `--force`），确认影响行数1和28。已有登记须先调查再重试。比较登记前后历史metadata和v3结果，不更新/删除历史行。

串行运行Go tests、race、vet、`scripts/test-production-v13.py` 和 `scripts/check-docs.py`。新overlay fixture验证127身份在双边选人及PVE/local/BC双座位准入、精确build/roster/鉴权拒绝、三组六英雄、锁定/请求重放、双边ready、对账及同房重赛。合成HTTP报告不代表浏览器/P2P通过。SQL检查使用本任务新建的可丢弃白名单socket，记录实际版本及跳过项。

构建使用 `scripts/build-profile.py --rosters deploy/production-v13/gameplay-rosters.json --features deploy/production-v13/protocol-features.json --goos linux --goarch amd64 --output bin/dueld-heros28`，记录两manifest及二进制SHA256。最终指纹复审后先增量SQL、后配套后端，再验证health、精确registry/runtime绑定及CORS。仅保留 `https://dotapk.lol` 和 `https://dotapk-frontend.skiyo.workers.dev`，不改TURN、代理、授权或网络。前端在后端就绪及真实PVP验收后单独协调发布，CPU不算PVP。回滚先撤新前端并排空28会话，再恢复旧后端/profile，保留metadata和历史。

## 历史profile

[gameplay-rosters.json](gameplay-rosters.json) 保留明确的 legacy20、旧 `arena-first22-46-v1` 和已发 `arena-heros22-v1` 绑定；不解锁当前前端46/127，前端仍22/88。默认 source embed 仍 legacy20。

固定复现工具输入：

| 输入 | 身份 |
| --- | --- |
| Go source commit | `aa04e01f25ef2226b213cfc64e68c6c96fd0e18c` |
| Profile SHA256 | `7fbdab24680a8b1b1101d5c596274e39a0bf757880fcf481e82da9ff5247023f` |
| 预期 Linux 二进制 SHA256 | `9512f02b72d20c97f14f8855559f42d902c6a13fb0fe8b7cd03baa859dcc1b8c` |
| 工具链 | Go1.26.1、CGO关闭、linux/amd64、已有模块cache |

根目录运行 `python3 scripts/build-production-v13.py` 会将实际源码与该提交比较，以 overlay 复现产物，不访问数据库/生产或重启服务。不一致需查原因，不改预期 digest；这是历史产物身份，不保证当前在线。

迁移004只初始化 legacy20，`register-roster.sql` 添加审核后的旧46，`register-heros22.sql` 添加22，append-runtime 文件保留审核后的精确 build列表。登记依赖匹配 schema 前置条件，INSERT 不能盲目重跑；运行 profile 与 SQL 元数据须一致，不用通配 build 或数组位置作为身份。

`heros22_profile_test.go` 由 `scripts/test-production-v13.py` 在 memory 下应用，不是游戏/浏览器测试。本地使用、运维审阅与回滚见[开发](../../docs/DEVELOPMENT.zh-CN.md)、[部署](../../docs/DEPLOYMENT.zh-CN.md)、[测试](../../docs/TESTING.zh-CN.md)。工作树不保留旧逐局/provision 回执，原 Git 历史仍在。

经审核的前端标准构建提交 `fcec6cf62c5ba53e397d3fef45e17ac9e859cb08` 使用 `duel-9431984810f197b393c5`，由 `append-privacy-release-runtime.sql` 精确追加绑定；既有22、46与legacy版本全部保留。英雄成员、规则和七字段质量策略不变。部署回执、主机配置、备份及逐局核验留在私有环境，不进入版本库。源码更新及这些产物摘要不表示前端已发布或浏览器验收通过。

新的进房选人源码会改变运行时代码字节，固定复现工具会明确拒绝。历史复现使用原源码 checkout，新候选用 [build-profile.py](../../scripts/build-profile.py) 显式指定审核后的 roster/features manifest，详见[选人发布边界](../../docs/SELECTION.zh-CN.md)。此现有 profile 保留全部旧审核构建。

审核后的 DOTA PK 前端73341e2运行时 `duel-2f81eeda15fb572139ad` 已精确追加到 released22 profile，也是该 profile 的 `protocol-features.json` 中 roomSelectionVersions 首个版本；保留全部旧绑定。`append-selection-runtime.sql` 用精确 compare-and-swap 更新既有 SQL 元数据，不新增迁移或权限。当前部署源码用 `scripts/build-profile.py` 同时指定两个生产 manifest 编译；此前固定摘要保留为历史身份。源码登记与前端实际发布分开。

DOTA PK 前端 `dad8e352` 使用精确运行时 `duel-2c2ad50b276ef596598c`。以 `append-materials-runtime.sql` 追加登记并与2f81一同启用选人，保留全部旧绑定与22/88规则；中间e81仅用于本地。当前2c2的HTTP/模块建房、取消、加入和锁定检查通过，未创建比赛；复用原e81自然P2P/重赛及独立native素材/fallback检查。因浏览器工具不可用，最终组合2c2 native双浏览器验收未执行，不能标为通过。无需新增后端逻辑、schema或权限。

DOTA PK 前端 `913f87ee` 使用最终运行时 `duel-64714553cccdaf8cc362`，22英雄与规则不变。`append-material-epoch-runtime.sql` 仅追加此最终身份，保留2f81/2c2，中间4a47仅限本地。最终HTTP加载前端模块通过真实API建房、加入、选人锁定与取消，未创建比赛或结果。Native WebRTC、控制包传输和已解码图像是显式模拟；真实浏览器/iOS冷热加载布局仍未验证。客户端素材证明校验当前epoch、英雄对与单调revision；后端API、schema和权限不变。
