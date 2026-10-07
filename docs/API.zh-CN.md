[English](API.md) | [简体中文](API.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# API 合约 v1.3

生产 base 为 `https://api.dotapk.lol/api/v1`，本地 `http://127.0.0.1:18082/api/v1`，JSON UTF-8、不用 cookie。GET `/healthz` 在 `/api/v1` 之外，探测 MySQL 并报告 `v1.3-gameplay-rosters`。Origin 必须与配置精确一致，生产为 `https://dotapk.lol`；OPTIONS 允许 GET/POST/DELETE/OPTIONS 和 Content-Type/Authorization。

POST `/sessions` body `{}` 返回 `{playerId,token,expires}`，256-bit ID/token，24小时过期。token 留本地，不放 URL/P2P/Git。除 registry/health/新会话外都需 `Authorization: Bearer <TOKEN>`；匿名会话不是持久人类身份。六位邀请码是保留前导零的字符串，不是强认证；内部房间/比赛 ID 是独立64位十六进制串。

## 注册表与精确版本

GET `/registry` 无 session 鉴权，仍受 origin/限流约束，返回 `{registryVersion,registrySha256,heroes,gameplayRosters}`。版本 `duel-heroes-127-v1`，SHA256 `5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138`，hash 是完整冻结 heroes 数组的 canonical 表示，HTTP/SQL 投影不能复算。英雄含 registryNumericId、internalHeroId、valveHeroId、valveHeroKey、legacyIndex，数字 ID 不是数组位置或 Valve ID。

普通 source embed 为 legacy20；[生产 profile](../deploy/production-v13/gameplay-rosters.json)保留 legacy20、旧46和已发22。当前 `arena-heros22-v1` 成员 `1,3,4,5,7,8,9,15,17,18,28,31,32,36,50,55,57,58,62,71,81,82`，精确绑定 `duel-27c78aa4cfc8facc8a23`、`duel-6b1d12f75aa4bbac4e12`、`duel-851e67d77307f479f1fa`。目录127不是 rosterId，其余105身份在该 build不可玩，旧46不解锁当前暂停英雄。

POST `/rooms`、`/rooms/join`、`/matches/pve`、`/matches/local` 支持可选非空字符串 rosterId，省略用 legacy20，显式空/null/错类型/未知名单为400；扩展 build必须显式发绑定名单，不符409、不用通配。响应保存 rosterId/registryVersion，旧缺失元数据投影 legacy20，不改原行。registryVersion/rulesHash 不可由客户端填请求，Go不执行技能，双方版本/名单一致。重试原 body保持不变：省略后改显式 roster会改变幂等 digest。

## 房间与信令

以下路由均相对 `/api/v1`。

| 请求 | 必填 body / 行为 |
| --- | --- |
| POST `/rooms` | `{requestId,version,hero,offer:{type:'offer',sdp},policy}`，另可选 rosterId，返回 id/code/expires/version/policy/players/offer |
| POST `/rooms/join` | `{code,version,hero}`，另可选 rosterId，原子占客人位；同客人可重试，其他客人409 |
| POST `/rooms/{id}/answer` | `{version,answer:{type:'answer',sdp}}`，answer不可变/幂等 |
| GET `/rooms/{id}` | 仅成员轮询，房主读 answer |
| DELETE `/rooms/{id}` | 任一成员关闭邀请/后续开局；不要仅因 SDP交换完成就关闭，保留房间以记录/再战 |

requestId 为16–80 ASCII字母数字/下划线/短横线，每操作/局新建、重试不变。policy 必填 direction `above|below`、rttMs1–2000、jitterMs30、lossPct5、minSamples24、window30、maxAgeMs3000；浏览器评估质量。hero必须在解析名单。入客前空位恰为 `{id:'',hero:0}`，不代表选择禁用英雄，仅房主房间准入接受该占位，比赛/客人不接受。

邀请码10分钟过期，复用代码不影响旧房间；30秒维护清过期 SDP，已连房间成员/再战至多24小时、token独立过期。关闭前上报进行中比赛中止，关房后已有比赛仍可接报告。无 WebSocket/战斗帧服务端传输，也无 TURN。

每分钟 IP/全局限额：全请求180/1200、新会话10/60、入房10/60、建房6/60，持久化 MySQL。用直接连接 IP，只有一个精确可信代理的合法 X-Real-IP 可替代，不信任任意 X-Forwarded-For。

## WebRTC 比赛生命周期

房主 POST `/rooms/{id}/matches` body `{requestId,version}` 返回 awaiting_ready，把比赛 ID通过可靠控制发客人。双方 POST `/matches/{id}/ready` body `{version}`，GET `/matches/{id}` 至 in_progress 才开战；ID作为 peer epoch。前局 live/pending 禁再战，终局才新 ID/requestId，客人不能建比赛。

双方独立上报观察到的历史，POST `/matches/{id}/results`，字段全部必填：

```json
{"version":"duel-build-hash","outcome":"completed","rounds":[{"number":1,"winner":0,"remainingMs":3400},{"number":2,"winner":0,"remainingMs":2000}],"score":[2,0],"winner":0,"reason":""}
```

这是合成 shape，不是真实记录。回合从1连续、最多64；winner0房主/1客人/-1平。remainingMs为客户端剩余秒×1000四舍五入，0–99000。score恰两值、一方第二胜后不能再有回合，完成 winner与先到2分一致。完整完成报告含剩余时间须一致，可靠投递最终观察，不指示 peer盲目确认他人的报告。

中止 body为 outcome aborted、winner-1、reason `left|disconnect|cancelled|version_mismatch`，只含已完成部分回合/两方<2分。报告不接受 player/hero ID、任意比赛/对手声明或客户端时间；服务端时间是收件时间，不是精确战斗遥测。

| 状态 | 含义 |
| --- | --- |
| awaiting_ready | 准备至多2分钟，否则 aborted/start_timeout |
| in_progress | 首报前至多20分钟，否则 aborted/result_timeout |
| pending | 一份有效报告，第二份等120秒，否则 aborted/result_timeout |
| confirmed | 双份完成报告一致，peer_agreement |
| disputed | 完成不一致或完成对中止，无赢家 |
| aborted | 中断/超时，不算正常胜局 |
| recorded | PVE/local/BC单个授权报告者完成，client_reported |

每座位报告不可变，相同重试成功、修改409、不延期限；HTTP隐藏原 submissions/digest，仅 reported双布尔。迟报不把超时升confirmed，200也可能已aborted，要看状态。

两份合法中止即使部分历史/原因不同也 aborted；相同原因保留，不同原因输出-only interrupted。completed+aborted → disputed/outcome_conflict，完成不一致 → disputed/conflicting_reports，与到达顺序无关。可选 scoreAgreement为相同比分 peer_agreed、单报 single_report、冲突/超时 unresolved；unresolved的 `[0,0]` 是占位不是平局，旧无字段表示未知。内部原始报告保持不可变。

## 人机、同屏与 BC

POST `/matches/pve` body `{requestId,version,hero,opponentHero,aiDifficulty:'easy'|'normal'|'hard'}`，另可选 rosterId，立即开局：人类seat0、AI seat1 ID为ai，不用AI token。前端当前normal，一份合法完成为 recorded/client_reported，中止仍中止。

POST `/matches/local` body `{requestId,version,hero,opponentHero,transport:'local'|'broadcastchannel'}`，另可选 rosterId，mode pvp、trust client_reported。两个服务端比赛内部 slot ID、participantKinds local_slot，不代表两个已鉴权的人；reporterPlayerId仅真实会话，只有创建者读取/上报，无AI或第二token。完成recorded，中止/超时aborted，不能请求confirmed/peer_agreement/webrtc。

BC仅房主创建/上报一条，把比赛/状态发客人共同显示，客人不重复建记录或假装独立认证双方。房主消失最终超时而非胜局，镜像状态是房主声明。API不可用时本地可保留待重试数据，UI不能说服务器已保存。

Match元数据含 id/roomId/version/players[{id,hero}]、mode/trust/ready/reported/status/score/winner/reason、时间/deadline、roster/registry及可选AI/transport/participant/reporter/agreement。WebRTC为两个anonymous_session，PVE anonymous_session+ai、local transport。旧行可能缺新字段，统计保留明确/推断历史cohort。

## 错误、统计与范围

400 shape/input、401会话、403成员/origin、404缺失、409座位/幂等/版本/状态、413 body>45KB、415媒体类型、429限流+Retry-After、503存储/服务。decoder拒未知/缺失/null字段与错误数组长度，MySQL不可用写失败，无D1/memory fallback。

迁移003增加独立v2 local/BC组，004增加身份/roster表和v3 view，不改旧记录/v1/v2含义。按build、roster/registry、mode/status/trust/transport、AI难度、英雄/对手/seat分组，PVP可每局两个appearance、PVE仅人类seat0；pending/disputed/aborted不计胜率，未映射/异常需人工质量审阅。没有公共统计/导出、排行榜、登录/管理UI、服务器模拟或反作弊。公开[平衡快照](https://github.com/dotapk-lol/heros/blob/main/balance-data/README.zh-CN.md)人工汇总审阅，不发原报告/标识。
