[English](SELECTION.md) | [简体中文](SELECTION.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# 进房选人与同房重赛

本次扩展已在源码实现，但此处没有启用新的生产前端构建。通过 `internal/duel/registry/protocol-features.json` 的 `roomSelectionVersions` 精确列表启用，默认发布列表为空。每个版本必须已绑定获准的可玩 roster，且包含默认英雄1；生成审核 overlay 时保留所有旧 roster/构建绑定。新构建不能通过省略选人请求或 `selectionEpoch` 绕过锁定。旧构建保持原 API 和请求摘要；PVE/local/BC 不变。

创建/加入房间以英雄1占位，沿用独立会话令牌和六位字符串邀请码。完成信令 answer 后，各自携带 Bearer 令牌调用 `POST /api/v1/rooms/{roomId}/selection`。要求双方会话有效、精确版本/roster 一致、房间未关闭且信令已回答。已连接房间从创建起最多24小时；后续选人/重赛无需重新消费或保留10分钟邀请。控制通道和20秒倒计时由前端负责。

## 开始选人与锁定

只有房主可开始。首次请求：

```json
{"version":"<获准精确构建>","action":"begin","epoch":"<新的UUID>","previousEpoch":"","previousMatchId":""}
```

下一轮的两个 previous 字段携带当前选人 epoch 和对应的已终结服务器 match ID。上一局须 confirmed/disputed/aborted，按已有终态与正常超时处理；active/pending 阻止开始。当前 epoch 完全相同请求重试返回当前视图，不重置锁定，即使建局已改变 `currentMatch`。修改请求、旧/重复使用的 epoch、错误上一轮关联返回409。房间有效期内不能回收旧 epoch。新 epoch 将房间双方英雄重置为1、锁定重置为 false，上一局快照与报告保持不可变。

双方分别锁定自己的认证座位：

```json
{"version":"<获准精确构建>","action":"lock","epoch":"<当前UUID>","hero":3}
```

英雄须属于精确获准 roster；未知字段、指定座位或对手英雄均拒绝。同英雄重试幂等，改已锁定英雄、解锁或旧 epoch 失败。为恢复丢失响应，建局后相同英雄的锁定重试仍可安全返回，不能修改已消费选人。房主不能代锁客人。成功响应后才能宣布 ready；倒计时到期也须双方各用自己的令牌请求。

两种操作及 GET room 返回原 RoomView 加可选 `selection:{epoch,previousEpoch,previousMatchId,locked:[bool,bool]}`；隐藏消费 match 标记和令牌哈希。新版本加入时以原英雄1占位的重试在本人选了其他英雄后仍有效。

## 建局与双方就绪

房主每轮使用新请求 ID：

```json
{"requestId":"<新一轮请求ID>","version":"<获准精确构建>","selectionEpoch":"<当前UUID>"}
```

要求双方服务器锁定、当前 epoch 尚未消费；原子消费 epoch，将实际锁定英雄快照写入 `match.players`。不同请求 ID 也不能为同一 epoch 再建一局，即使已终结。完全相同重试返回同一局，修改请求返回409。创建/读取/ready/结果视图回显可选 `selectionEpoch`，省略字段保持旧 JSON/摘要。双方仍各调用 `/matches/{id}/ready`，正文 `{version}`；仅双方确认后进入 `in_progress`。结果正文、对账与可信度分组不变。重赛保留房间/会话/P2P，结束上一局报告后，新 epoch、重新锁定、新 match。

## Casual 网络参数

显式启用的新构建可使用唯一新增的七字段参数：

```json
{"direction":"above","rttMs":500,"jitterMs":250,"lossPct":30,"minSamples":1,"window":12,"maxAgeMs":10000}
```

旧客户端继续接受旧校验规则；新增 tuple 精确回显，任意变体拒绝。参数用于浏览器警告/采样展示，不是后端质量否决条件。后端不等待24样本，不模拟网络或战斗帧。浏览器区分实时提示与真正的可靠控制断连/严重积压。未声称物理5G验收或竞技级防作弊；串通的 P2P 双方仍可伪造一致记录。

## 候选构建与发布边界

显式指定审核后的 roster/features manifest，脚本使用临时 Go overlay、缓存依赖和受限并发编译当前源码：

```sh
python3 scripts/build-profile.py --rosters '<审核ROSTERS_JSON>' --features '<审核FEATURES_JSON>' --output bin/dueld-candidate
```

部署制品加 `--goos linux --goarch amd64`。脚本不访问数据库、不启动服务、不发布。旧 `build-production-v13.py` 仍是历史固定制品复现脚本，会拒绝已改变的运行时代码；回滚复现须使用原源码 checkout，不得修改旧摘要来掩盖新代码。

选人/局关联存在已有 JSON payload，无需修改 migration1–4、历史记录或运行时数据库权限。启用前须冻结精确前端运行时，在两个 manifest 及已有 SQL roster 版本元数据绑定，通过隔离 API/浏览器验收后，再协调现有 API 监听与 origin 部署；不增加公网监听或创建凭据。回滚恢复旧二进制/profile，历史 JSON 可保留新增字段。旧服务器不执行新选人门槛，因此新前端也须同时回滚或禁用。本地 QA 标签不授权任何生产构建。

审核后的 DOTA PK 前端73341e2运行时 `duel-2f81eeda15fb572139ad` 已精确追加到 released22 profile，也是该 profile 的 `protocol-features.json` 中 roomSelectionVersions 首个版本；保留全部旧绑定。`append-selection-runtime.sql` 用精确 compare-and-swap 更新既有 SQL 元数据，不新增迁移或权限。当前部署源码用 `scripts/build-profile.py` 同时指定两个生产 manifest 编译；此前固定摘要保留为历史身份。源码登记与前端实际发布分开。

DOTA PK 前端 `dad8e352` 使用精确运行时 `duel-2c2ad50b276ef596598c`。以 `append-materials-runtime.sql` 追加登记并与2f81一同启用选人，保留全部旧绑定与22/88规则；中间e81仅用于本地。当前2c2的HTTP/模块建房、取消、加入和锁定检查通过，未创建比赛；复用原e81自然P2P/重赛及独立native素材/fallback检查。因浏览器工具不可用，最终组合2c2 native双浏览器验收未执行，不能标为通过。无需新增后端逻辑、schema或权限。
