[English](API.md) | [简体中文](API.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# API contract v1.3

Production base: `https://api.dotapk.lol/api/v1`; local base: `http://127.0.0.1:18082/api/v1`. JSON UTF-8, no cookies. GET `/healthz` is outside `/api/v1`, probes MySQL and reports `v1.3-gameplay-rosters`. An Origin header must exactly match configured origin; production uses `https://dotapk.lol`. OPTIONS allows GET/POST/DELETE/OPTIONS and Content-Type/Authorization.

POST `/sessions` body `{}` returns `{playerId,token,expires}` with256-bit identifiers/tokens and24-hour expiry. Keep tokens local, outside URLs/P2P/Git. Except registry/health/session creation, routes require `Authorization: Bearer <TOKEN>`. Anonymous sessions are not durable human identities. A six-digit invitation string preserves leading zeros; it is not strong authentication. Internal room/match IDs are independent64-character hex values.

## Registry and exact versions

GET `/registry` returns `{registryVersion,registrySha256,heroes,gameplayRosters}` without session auth, still subject to origin/rate limits. Registry is `duel-heroes-127-v1`, SHA256 `5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138`. The digest covers the canonical full frozen heroes array; projected HTTP/SQL fields cannot reproduce it. Hero fields include registryNumericId, internalHeroId, valveHeroId, valveHeroKey, legacyIndex. Numeric ID is not array position or Valve ID.

Ordinary source embed admits legacy20. [Production profile](../deploy/production-v13/gameplay-rosters.json) preserves legacy20, historical46 and released22. Current `arena-heros22-v1` uses IDs `1,3,4,5,7,8,9,15,17,18,28,31,32,36,50,55,57,58,62,71,81,82`, exactly bound to `duel-27c78aa4cfc8facc8a23`, `duel-6b1d12f75aa4bbac4e12`, `duel-851e67d77307f479f1fa`. Catalog127 is not a rosterId; other105 identities are not playable in that build. Historical46 does not unlock paused heroes in the current frontend.

POST `/rooms`, `/rooms/join`, `/matches/pve`, `/matches/local` accept optional nonempty string `rosterId`; omission is legacy20, explicit empty/null/wrong type/unknown roster is400. Expanded build requires its explicit bound roster (409 on mismatch), no wildcard. Responses snapshot rosterId/registryVersion; older missing metadata projects as legacy20 without row rewrites. registryVersion/rulesHash are not client request fields; Go does not execute skills. Both peers need the same exact version/roster. Retry the original unchanged body: adding an explicit roster to a previously omitted one changes the idempotency digest.

## Rooms and signaling

All route names below are relative to `/api/v1`.

| Request | Mandatory body / behavior |
| --- | --- |
| POST `/rooms` | `{requestId,version,hero,offer:{type:'offer',sdp},policy}` plus optional rosterId; returns room with id/code/expires/version/policy/players/offer |
| POST `/rooms/join` | `{code,version,hero}` plus optional rosterId; atomically claims guest, same guest may replay, another guest409 |
| POST `/rooms/{id}/answer` | `{version,answer:{type:'answer',sdp}}`; immutable/idempotent answer |
| GET `/rooms/{id}` | Participant-only polling, host reads answer |
| DELETE `/rooms/{id}` | Either participant closes invitations/future starts; do not close merely after SDP exchange, retain room for results/rematches |

requestId is16–80 ASCII alphanumeric/underscore/hyphen, new per operation/game, unchanged on retry. policy requires direction `above|below`, rttMs1–2000, jitterMs30, lossPct5, minSamples24, window30, maxAgeMs3000; browser evaluates quality. Hero must be in the resolved roster. Empty guest is exactly `{id:'',hero:0}` before joining, not an excluded hero selection; only host room admission accepts that placeholder, not match/guest admission.

Invitation expiry is10 minutes, independent reused codes cannot affect old rooms.30-second housekeeping erases expired SDP; connected room membership/rematches last at most24 hours, tokens expire separately. Report an active game's abort before closing. Existing matches may still receive reports after room closure. There is no WebSocket or combat-frame server transport and no TURN.

Limits per minute (IP/global): all requests180/1200, session creation10/60, joins10/60, room creation6/60. Limits persist in MySQL. Direct socket IP is used unless one exact trusted proxy supplies valid X-Real-IP; arbitrary X-Forwarded-For is not trusted.

## WebRTC match lifecycle

Host POST `/rooms/{id}/matches` body `{requestId,version}` returns `awaiting_ready`. Host sends the returned match ID over reliable control. Each participant POST `/matches/{id}/ready` body `{version}`, then GET `/matches/{id}` until `in_progress`; start only then. The ID is the peer epoch. A live/pending previous game blocks rematch; a terminal game permits a new ID/requestId. Guest cannot create a match.

Both clients independently submit their observed histories via POST `/matches/{id}/results` with every field mandatory:

```json
{"version":"duel-build-hash","outcome":"completed","rounds":[{"number":1,"winner":0,"remainingMs":3400},{"number":2,"winner":0,"remainingMs":2000}],"score":[2,0],"winner":0,"reason":""}
```

This is synthetic shape data, not a real record. Round numbering starts1 and is contiguous, max64; winner0 host/1 guest/-1 tie. remainingMs is rounded client remaining seconds×1000, bounded0–99000. score has exactly2 values and no rounds follow either side's second win; completed winner matches first-to2 score. Full completed reports, including remaining time, must agree. Deliver final observations reliably, never instruct a peer to blindly affirm another report.

Abort body uses outcome `aborted`, winner-1, reason `left|disconnect|cancelled|version_mismatch`, completed partial rounds/derived score below2. No player/hero IDs, arbitrary match/opponent assertions or client timestamps are accepted in reports. Server times are receipts, not precise combat telemetry.

| State | Meaning |
| --- | --- |
| awaiting_ready | Up to2 minutes, then aborted/start_timeout |
| in_progress | Up to20 minutes before first report, then aborted/result_timeout |
| pending | One valid report; second has120 seconds, then aborted/result_timeout |
| confirmed | Two identical completed reports, trust peer_agreement |
| disputed | Completed disagreement or completed vs aborted, no winner |
| aborted | Interruption/timeout, no normal win |
| recorded | Valid PVE/local/BC completion from one authorized reporter, client_reported |

Per-seat reports are immutable: identical retries succeed, changed retries409, no deadline extension. HTTP hides raw submissions/digests and exposes `reported:[bool,bool]`. Late results cannot promote timeout to confirmed; a200 response may already be aborted, so inspect status.

Two valid aborted reports resolve aborted even with different observed partial histories/reasons; matching reason is preserved, differing reason becomes output-only `interrupted`. completed+aborted → disputed/outcome_conflict; nonidentical completed reports → disputed/conflicting_reports, independent of arrival order. Optional `scoreAgreement` is peer_agreed for matching scores, single_report for self-report, unresolved for conflicts/timeouts. `[0,0]` with unresolved is a placeholder, not a draw. Old rows without this field are unknown. Original reports remain immutable internally.

## PVE, local and BC

POST `/matches/pve` body `{requestId,version,hero,opponentHero,aiDifficulty:'easy'|'normal'|'hard'}` plus optional rosterId starts immediately: human seat0, AI seat1 ID `ai`, no AI token. Frontend currently uses normal difficulty. One valid result is recorded/client_reported, abort stays aborted.

POST `/matches/local` body `{requestId,version,hero,opponentHero,transport:'local'|'broadcastchannel'}` plus optional rosterId starts mode pvp, trust client_reported. Two server-generated match-local slot IDs with participantKinds local_slot do not assert two authenticated people. reporterPlayerId is the one real session; only creator reads/submits, no AI or second token. Completion recorded, abort/timeout aborted; callers cannot request confirmed/peer_agreement/webrtc here.

BC host alone creates/submits one record, sending match/status to guest for shared display; guest does not create a duplicate or separately authenticate participants. Missing host eventually times out, not a win. Mirrored status is a host assertion. Local browser play may retain unsaved retry data when API fails; UI must not label it server-saved.

Match metadata includes id/roomId/version/players[{id,hero}], mode, trust, ready/reported/status/score/winner/reason, timestamps/deadline, roster/registry, optional AI/transport/participant/reporter/agreement fields. WebRTC kinds are two anonymous_session; PVE anonymous_session+ai with local transport. Historical rows may omit newer fields; analytics retains explicit/inferred historical cohorts.

## Errors, analytics and scope

400 shape/input,401 session,403 membership/origin,404 missing,409 seat/idempotency/version/state,413 body>45KB,415 media type,429 limit with Retry-After,503 storage/service unavailable. Decoder rejects unknown/missing/null fields and wrong array lengths. Runtime writes fail if MySQL is unavailable, no D1/memory fallback.

Migration003 adds separate v2 local/BC cohorts;004 adds hero/roster tables and v3 views without rewriting old records/v1/v2 meanings. Group build, roster/registry, mode/status/trust/transport, AI difficulty, hero/opponent/seat; PVP appearances may be two per game, PVE human seat0 only. Pending/disputed/aborted do not enter win rates; unmapped identities and anomalies need manual quality review. No public analytics/export, leaderboard, login/admin UI, server simulation or anti-cheat is supplied. Public [balance snapshots](https://github.com/dotapk-lol/heros/blob/main/balance-data/README.md) are manually aggregated/reviewed, never raw reports or identifiers.

Current source adds exact-build-gated [room-first selection and same-room rematches](SELECTION.md). Default feature bindings are empty; this is not a production activation. Memory/race and disposable MySQL tests cover epoch replay, own-seat locks, immutable prior matches, simultaneous locks, single allocation, expiry and legacy request bytes.
