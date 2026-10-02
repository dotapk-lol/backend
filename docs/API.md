# Frontend integration contract v1

Production base `https://api.dotapk.lol/api/v1`, JSON UTF-8, no cookies. CORS origin is exactly `https://dotapk.lol`. `POST /sessions` issues `{playerId,token,expires}` (server 256-bit IDs/tokens, 24 hours). Store token locally; never send it over the P2P channel or put it in URLs. All other endpoints use `Authorization: Bearer <token>`. Each browser gets a separate anonymous session. This is not a login or durable human identity. UI must state when results cannot be saved; never show local wins as backend-confirmed.

Room invitation `code` is **exactly six ASCII digits as a string** (including `000007`). It is a short invitation, not strong authentication. Internal room/match IDs are independent 64-character hex strings. Anonymous player identity is bound to the session hash and occupied seat; results do not accept arbitrary player IDs or heroes.

## Signaling

1. `POST /rooms` body `{requestId,version,hero,offer:{type:'offer',sdp},policy}` -> room with `id`, `code`, `expires` (Unix milliseconds), `version`, `policy`, `players`, `offer`. `requestId`: random 16–80 ASCII alphanumeric/underscore/hyphen; retain unchanged for retries. hero is an integer 0–19. `policy` requires all existing fields: direction `above|below`, rttMs 1–2000, jitterMs 30, lossPct 5, minSamples 24, window 30, maxAgeMs 3000. The server stores the existing policy semantics; browser still evaluates quality.
2. `POST /rooms/join` body `{code,version,hero}` atomically claims guest slot and returns room/offer. Same guest may replay; another guest gets 409. Version must exactly match.
3. Guest `POST /rooms/{id}/answer` body `{version,answer:{type:'answer',sdp}}`. Answer immutable/idempotent.
4. Host `GET /rooms/{id}` polls `answer`. Only the two room sessions can read. Do not use the invitation code in internal routes.
5. `DELETE /rooms/{id}` closes invitations and prevents future starts; either participant may leave. **Do not call this after SDP exchange** as the old client did. Keep room membership for results/rematches. Existing matches may still receive reports after room closure. Send an aborted report for a live game before closing.

Codes expire after 10 minutes; reused codes cannot affect earlier rooms. SDP is purged by a 30-second maintenance task after expiration. Connected room membership/rematches last at most 24 hours; tokens expire independently. Joining consumes 10 attempts/IP/minute and 60 globally/minute. New rooms 6/IP/minute and 60 global, session creation 10/IP/minute and 60 global, all requests 180/IP/minute and 1200 global. Limits persist in MySQL. Direct socket IP is used; optional one trusted loopback proxy may set sanitized X-Real-IP. Never trust arbitrary X-Forwarded-For. Internet-scale distributed brute-force resistance is outside this prototype.

## PVP match lifecycle

Host `POST /rooms/{id}/matches` body `{requestId,version}` -> Match, initially `awaiting_ready`. Every replay/rematch uses a NEW requestId. Retry the SAME requestId after lost response. The host sends returned `id` to its peer over reliable control. Both clients `POST /matches/{id}/ready` with `{version}` then poll `GET /matches/{id}` until status `in_progress`; only then start fighting. Use this server match ID as the P2P epoch. Server timestamps start when both ready. A previous live/pending match blocks rematch creation; confirmed/disputed/aborted permits a fresh match ID. Guest cannot create matches.

At match end **both clients independently submit their observed history**, including tied rounds. `POST /matches/{id}/results` body (all fields mandatory):

```json
{"version":"duel-build-hash","outcome":"completed","rounds":[{"number":1,"winner":0,"remainingMs":3400},{"number":2,"winner":0,"remainingMs":2000}],"score":[2,0],"winner":0,"reason":""}
```

Map engine history with `number=round`, `remainingMs=Math.round(remaining*1000)`, `winner` 0 host / 1 guest / -1 draw. score is exactly 2 elements; round order begins at 1, is contiguous, ≤64 rounds, no rounds after either side gets 2 wins. Winner must match the computed first-to-2 score. For completed games, the full semantic report, including remaining time, must match. Host should send the final snapshot reliably so guest observes the same canonical terminal history; never send instructions for guest to blindly affirm a result it did not observe.

For exit/disconnect use `outcome:'aborted'`, `winner:-1`, reason `left|disconnect|cancelled|version_mismatch`, partial completed rounds and derived score (<2 each). A departure never awards a normal win. Local time, player IDs, hero IDs, arbitrary match IDs, and opponent claims are not accepted in reports. Server stores receive times and match start/end times (receipt-based, not precise combat-duration telemetry).

States:

- `awaiting_ready`: ≤2 minutes to start; then `aborted/start_timeout`.
- `in_progress`: ≤20 minutes to first result; then `aborted/result_timeout`.
- `pending`: only one valid report; second has 120 seconds; then `aborted/result_timeout`.
- `confirmed`: two identical completed reports.
- `disputed`: completed reports differ, or completed is contradicted by aborted; no winning result is awarded.
- `aborted`: both report interruption (normal local reasons may differ), or timeout; exclude from normal win rates.

A report is immutable per seat: exact retries succeed, altered retries 409. GET/POST responses do **not** expose raw submissions or hashes (prevent copying the opponent report through HTTP); `reported:[bool,bool]` indicates receipts. Confirmed state is immutable. Late results cannot turn timed-out games into confirmed games. A submit that crosses deadline may return 200 with `aborted`; always inspect status.

## PVE

`POST /matches/pve` with `{requestId,version,hero,opponentHero,aiDifficulty:'easy'|'normal'|'hard'}` starts one match immediately. Server registers human seat 0, AI seat 1 with ID `ai`; no AI token. Human submits the same results shape once. Valid completion becomes `recorded`, `trust:'client_reported'`, never PVP `confirmed`. Abort stays aborted. The browser should set difficulty to the implemented AI level (currently normal unless explicitly changed). PVE retries use the same requestId; each fresh game a new requestId.

Match responses include id, roomId, version, players[{id,hero}], mode, aiDifficulty (PVE), trust, ready, reported, status, score, winner, reason, createdAt/startedAt/endedAt/deadline (Unix ms). PVP trust `peer_agreement` means both clients agreed; colluding P2P clients can still fabricate games. No competitive-grade anti-cheat is claimed. PVE has only client assertion.

## Errors and integration boundaries

400 invalid/missing JSON fields, 401 invalid/expired session, 403 membership/origin, 404 unavailable, 409 seat/request/version/state conflict, 413 >45KB, 415 media type, 429 limit (+Retry-After), 503 storage/service unavailable. Strict input decoding rejects extra fields, missing fields, nulls, wrong score length. GET `/healthz` probes MySQL, no auth.

All signaling/results move from Worker+D1 to this API. No combat frames go to Go. No TURN, leaderboard, login or admin UI is created. Approved api.dotapk.lol DNS and dedicated TLS now route the public API through existing Nginx443 to loopback Go18082; frontend root dotapk.lol is on Cloudflare. The older private Sites origin is not on the production CORS allowlist. Do not silently fall back to D1 or label a failed server write as saved.

## Local and BroadcastChannel PVP extension (v2 candidate)

`POST /matches/local` requires the real reporting session's Bearer token and all fields:

```json
{"requestId":"new-random-id-per-game","version":"duel-build-hash","hero":0,"opponentHero":3,"transport":"local"}
```

`transport` is exactly `local` (same-screen) or `broadcastchannel` (same browser windows). Starts immediately with `mode:'pvp'`, `trust:'client_reported'`, `status:'in_progress'`. `players` contain two independent server-generated **match-local slot IDs**, each `participantKinds:'local_slot'`; these do not assert two authenticated people. `reporterPlayerId` identifies the one actual authenticated reporter. There is no AI participant, AI difficulty or second token.

Submit the same `/matches/{id}/results` shape. Completion becomes `recorded`; abort/timeout becomes `aborted`. Only the creator's session may read/submit. Identical create/result retries are idempotent; altered retries conflict. Every rematch needs a fresh requestId and receives a new match and new local slot IDs. Callers cannot request `confirmed`, `peer_agreement`, a different reporter or a `webrtc` transport through this endpoint.

For BroadcastChannel, the host alone creates/submits one server record. Send the returned match ID and recording status over BC to the guest for a shared/deduplicated display. The guest does not create a second record or pretend to independently authenticate both players. This choice matches the existing BC host-authoritative simulation and makes trust explicit. If either window exits, host reports the observed abort when able; an unavailable host eventually yields a timeout, not a normal win. Guest-reported mirrored UI status is only the host's assertion; it is not separate server authentication.

Frontend `recorded` copy must be generic: “战绩已保存 · 客户端上报”, rather than “人机战绩已保存”. Local playback may continue when the backend is unavailable, but UI must clearly say unsaved and retain retry data; never claim the browser-only record reached MySQL.

New P2P matches also expose `transport:'webrtc'` and `participantKinds:['anonymous_session','anonymous_session']`; PVE exposes `transport:'local'`, kinds `['anonymous_session','ai']`. Existing P2P/PVE routes and report payloads are unchanged. Old rows may lack these newly added fields; SQLv2 infers their transport based on the only previously supported server modes.

Migration003 adds `duel_hero_balance_v2` and `duel_data_quality_v2` without updating old rows or replacing legacy views. The v2 balance view includes both slots for completed local/BC PVP, grouped separately by **version, mode, transport, trust, status, hero, opponent hero and seat**. Never merge peer-confirmed and client-reported cohorts by dropping these dimensions. Aborted/disputed/pending games still do not enter normal win rates. The original `duel_hero_balance` keeps its existing meaning and excludes local/BC self-reports.

## Abort reconciliation fix (v1.2)

Health contract version: `v1.2-abort-reconciliation`. Inputs/routes unchanged.

Two independently valid `aborted` reports now resolve to `aborted` with winner `-1` even when local viewpoints differ (`left` vs `disconnect`, or different last-observed partial rounds). Each original report, digest and receivedAt remains immutable in MySQL. Matching reasons are preserved; differing reasons yield output-only summary reason `interrupted` (not a new accepted client input value).

The summary exposes optional `scoreAgreement`: `peer_agreed` when both scores match (also true for identical completed reports), `single_report` for recorded/aborted local or PVE self-reports, and `unresolved` for conflicts/timeouts. For two aborted reports with different partial scores, summary score `[0,0]` is only a placeholder with `scoreAgreement:'unresolved'`, not a claimed draw; original partial histories remain in the two stored submissions. Client UIs must show the match as interrupted, not declare a winner from partial scores. Older rows without this additive field have unknown summary-agreement metadata.

| First and second reports | Summary |
| --- | --- |
| aborted + aborted (same or different reason/history) | aborted; no winner |
| completed + aborted, in either arrival order | disputed / outcome_conflict; no winner |
| completed + completed, full reports identical | confirmed |
| completed + completed, any substantive result difference | disputed / conflicting_reports; no winner |
| only one report before timeout | aborted / result_timeout; no winner |

Arrival order does not change the resolution. Exact retries do not extend pending deadlines or change receipts. Altered retries remain rejected. A late second report cannot reopen a timed-out match. Aborted/disputed matches remain excluded from all normal win-rate cohorts. Previously stored terminal QA records are left unchanged for audit; this update changes future reconciliation only.
