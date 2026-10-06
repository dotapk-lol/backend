# DotaPK production deployment

Current backend state (2026-10-06): **The existing22 roster now retains27c and additionally accepts the exact6b room-validation build, preserving every old roster/build. Only its version list and the Go embedded profile changed; DOTA DUEL alone restarted.** No existing roster, hero identity, schema, credential, grant, Nginx route or other application changed. Frontend publication/browser acceptance belongs to the parent task. See the current runtime receipts below; October3 receipts remain historical.

## Repository and runtime

- Private repository: https://github.com/dotapk-lol/backend
- Host: existing `a2-webserver`, `43.162.87.40`, x86_64, OpenCloudOS,2 CPUs, about6.2GB available memory/69GB free disk at inspection.
- Game service remains `dota-duel.service`, created during the authorized initial deployment and reused for this upgrade. No new MySQL instance, Docker, Redis, TURN, proxy daemon, certbot scheduler or other service was added.
- Binary `/opt/dota-duel/releases/v1.3-9962a885-room-fix/dueld`; `/opt/dota-duel/bin/dueld` symlink. Linux SHA256 `9962a885ccd42ee8cc8fbd878bef1724b552c35906c8daaca71f4f2536e3c0d3`. Source snapshot `aa04e01f25ef2226b213cfc64e68c6c96fd0e18c` plus `deploy/production-v13/gameplay-rosters.json`, Go1.26.1; `scripts/build-production-v13.py` reproduces and checks the exact binary locally.
- Production profile retains legacy20 and `arena-first22-46-v1` (e63/7e), retaining `arena-heros22-v1` for exact `duel-27c78aa4cfc8facc8a23` and `duel-6b1d12f75aa4bbac4e12`. No catalog127/CORE4-24/later candidate activation. A generic default build still enables only legacy20.
- Go binds `127.0.0.1:18082` only. systemd DynamicUser, MemoryMax128MiB, CPUQuota25%, TasksMax32; DB pool5 open/2 idle. User authorized the server deployment and dedicated identity.
- Exact CORS origin `https://dotapk.lol`; only trusted proxy `127.0.0.1`. Nginx overwrites X-Real-IP. Keep API DNS-only unless Cloudflare proxy client-IP trust is separately configured; never blindly trust incoming CF/X-Forwarded-For headers.

## Database isolation

Existing MySQL8.4.8 at127.0.0.1:3306 is reused. Dedicated schema `dota_duel`, migrations1/2/3/4 applied and tested. Migration004 adds3 catalog/roster tables and4 v3 views; `deploy/production-v13/register-roster.sql` separately registers approved46 membership. Runtime DML grants are unchanged. Existing business schemas and users were not modified. Schema separation is database permission isolation, not separate hardware/process or protection from the host administrator.

Inspection showed `agentsquared@127.0.0.1` has ALL PRIVILEGES on agentsquared_website, so it was not reused. With explicit user approval, created only `duel_app@127.0.0.1` with SELECT, INSERT, UPDATE, DELETE on dota_duel.*. Runtime login/read passed; reads of mysql.user and the AgentSquared business table were independently denied. No global, DDL, grant or other-schema privileges.

A32-byte CSPRNG secret was generated only on the server, never displayed or committed. `/etc/dota-duel/mysql-dsn` is root:root0600 under0700 directory; systemd LoadCredential exposes it to Go. Public non-secret origin settings are separate. Existing root administrator authentication was used only for authorized schema/provisioning operations; the game does not receive root or A2 business credentials. `deploy/provision-database.py` refuses existing identity/secret overwrite. Do not rerun it casually. `deploy/check-runtime-access.py` is a read-only verification helper.

## Nginx and certificate plan

Existing Nginx1.26.3, Certbot2.8 webroot `/var/www/certbot`, and existing `/etc/cron.d/agentsquared-certbot-renew` are reused. That existing cron renews daily at03:17 and reloads nginx through a deploy hook. No new renewal service/timer is required.

`api.dotapk.lol` requires its own certificate/key; AgentSquared/Pikoo certificates are not valid substitutes. DNS-only A=43.162.87.40 was verified publicly and on the host, with no AAAA. HTTP-01 succeeded over the real domain. Existing ACME account issued a dedicated api.dotapk.lol certificate (Let's Encrypt YE2, expires2026-12-31); its private key is root:root0600 and never leaves the server. The443 vhost is installed after backup and nginx -t, followed by reload. Its only proxied routes are `/api/v1/` and `/healthz` to127.0.0.1:18082; no new public port is opened. Root frontend remains on Cloudflare.

Before changes, existing Nginx conf.d was copied to `/var/backups/dota-duel/initial-20261002/nginx-conf.d`; baseline hashes are in existing-nginx.sha256. A2/Pikoo files remain byte-for-byte unchanged after HTTP-stage and final443 reloads. Existing A2 healthz and Pikoo /health/live plus /health/ready returnedok; nginx/A2/MySQL/Redis stayedactive. A pre-existing unrelated tat_agent unit warning was observed, not modified.

## October6 room-validator runtime append

- Frontend1affd6d fixes room-stage validation of Go empty guest slots; no Go logic, hero selection or policy changes. Only `arena-heros22-v1.game_versions` appends6b with compare-and-swap SQL (`append-room-fix-runtime.sql`), exactly1 row affected. Old27c,46/legacy rosters and every member remain unchanged.
- Protected dedicated backup `/var/backups/dota-duel/upgrade-v13-20261006T110504Z`, SHA256 `d1859ca6a1cf6ba9fd85f60ab0551e658136aebb7d0e1a99132c4b0e0ee6d2b0`; retrieved local0600 copy verified. No dump/credential committed.
- Pinned Go profile tests and exact binary reproduction pass. Only DOTA DUEL restarted, 0.19s to loopback ready; existing apps/configuration/grants/credentials unchanged.
- 34 external HTTPS checks: actual new host RoomView with `{id:"",hero:0}` passes the dedicated validator; occupied joined seats pass, strict match validation still rejects the empty slot. Old27c/46 and legacy signaling remain accepted; wrong roster/build/hero requests rejected. Four QA rooms closed, no match created.
- All35 existing payload hashes,127 identities,88 memberships and8 views preserved. Evidence: `room-fix-deployment.json`, `room-fix-https-validator-evidence.json`, `room-fix-sql-evidence.json`, `room-fix-build-reproduction.json`, `room-fix-profile-tests.txt`. Validator checks use precommit emitted modules described by the frontend handoff; the parent must rebuild exact1affd6d before Cloudflare publication. No native browser/WebRTC or combat acceptance is claimed.

## October6 independent22 release (historical)

- Frozen frontend4608b48 and all36 handoff files SHA-verified; independent review SHA4b3019ae verified. Exact22 IDs and sole27c build come from that freeze; registry127 and all prior mappings are unchanged.
- Dedicated protected backup `/var/backups/dota-duel/upgrade-v13-20261006T081134Z`, SHA256 `27d717948d482b970b8817271974c74a6f3be8a2aa7a42e1222cdf44d6d541b9`; local0600 copy verified. No dump or credential is committed.
- `register-heros22.sql` transaction inserts1 independent roster and22 membership rows, without an upsert or old-row update. This is additive DML using the existing operator, not a new migration/privilege.
- Only Go restarted;0.189s to loopback ready. Protected configuration, runtime grants and credentials, existing service PIDs and A2/Pikoo health unchanged.
- Serial in-memory tests cover all127 IDs at room/join/PVE/local/BC boundaries, old request bytes and result/rematch behavior.37 external TLS/CORS/signaling/rejection checks pass. Three rooms closed; no match created. Actual22 built modules match frozen archive (93 JS files); their registry validator and both old46 validators return verified.
- All34 payload hashes,127 identities, old66 membership rows and8 view definitions retained; new22 membership exact. Migrations still1–4.
- Current evidence: `heros22-deployment.json`, `heros22-https-evidence.json`, `heros22-candidate-registry.json`, `heros22-sql-evidence.json`, `heros22-build-reproduction.json`, `heros22-profile-tests.txt`, `heros22-freeze-verification.json`, `heros22-review-verification.json`. These are backend/API checks; no natural browser/device/WebRTC or skill acceptance is claimed.

## October6 exact UI runtime registration (historical)

- Fresh protected dedicated backup: `/var/backups/dota-duel/upgrade-v13-20261006T061701Z/`; SHA256 `598af630894f6ef9caa93ad0797ab11b2974da8d16a5706c235ca1e659dd4f2d`. Local protected copy was also SHA-verified; dumps are excluded from this repository.
- `deploy/production-v13/append-runtime-7e767a8.sql` is a one-time compare-and-swap DML append, not a new migration or seed replacement. Exactly1 row changed, preserving e63 and the exact46 membership. Do not rerun blindly.
- Only Go restarted,0.188s to loopback ready, PID151973/NRestarts0 at this check. Existing service PIDs, A2/Pikoo health, all protected configuration and credential/grants remained unchanged.
- 29 external TLS/CORS/signaling checks passed across old46, new46 and omitted-roster legacy20; mixed-version joins rejected. Three signaling rooms were closed, no matches created. Actual frontend43bf built compatibility validator accepted both exact builds and all46 IDs. No browser RTT/quality or combat acceptance is claimed.
- All34 existing match payload hashes,127 identities,66 membership rows and8 view definitions were preserved; migrations remain1–4, active matches0 at handoff.
- Current evidence: `runtime-7e767a8-deployment.json`, `runtime-7e767a8-https-evidence.json`, `runtime-7e767a8-candidate-registry.json`, `runtime-7e767a8-sql-evidence.json`, `runtime-7e767a8-build-reproduction.json`. The immutable server build metadata records build facts; repository metadata additionally records successful production application.

## October3 v1.3 production release acceptance (historical)

- No active match before restart. Protected dedicated backups retained before target8.4 migration (`/var/backups/dota-duel/upgrade-v13-20261003T025426Z/`) and before Go switch (`/var/backups/dota-duel/upgrade-v13-20261003T030233Z/`). Full dumps are not in this repository. An earlier dedicated backup was actually restored into a separate local9.6 socket, reproducing all7 original table fingerprints and rehearsing the additive migration; this did not replace target8.4 verification.
- Actual MySQL8.4.8 executed004 and46 registration. Original22 match payloads/29 immutable report digests and old4 view definitions unchanged; all127 SQL mappings equal the frozen identity projection. Membership20+46 and sole46 binding verified. The existing service continued normal expiry housekeeping of session/request references; match/report evidence was retained.
- Only Go restarted:0.144s from restart command to loopback health (not a measurement of every client's outage). PID2684766/NRestarts0 at handoff. Nginx/MySQL/Redis PIDs, vhost hashes, unit/public.env and credential content/permissions unchanged; A2 health ok and Pikoo live ok/ready ready before and after.
- External TLS verification, exact CORS/OPTIONS, foreign origin rejection, old20 omitted-roster/signaling/peer/rematch/local+BC paths, new46 PVE/PVP abort and rejected inactive/wrong-bound/unsupported selections:49 checks passed. No session tokens or raw submissions appear in the new evidence.
- Final SELECT-only verification: original22 payloads/29 reports unchanged,7 synthetic HTTP-only games/10 reports verified,29 total games and0 active at that instant; all v3 integrity counters0. Only one synthetic completion exists under `qa-v13-legacy20`; both46 tests aborted/no winner and do not contribute wins. Filter these synthetic rows from natural browser cohorts.
- Evidence: `production-v13-deployment.json`, `production-v13-ready.json`, `production-v13-https-evidence.json`, `production-v13-sql-evidence.json`. Backend/API checks do not claim natural46 browser combat or competitive anti-cheat.

Full frozen-array digest `5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138` includes allocation metadata/order omitted by SQL/public projections. SQL's five identity fields ordered by numericID separately hash to `1b2f96b9493172b9d58603c6d2c5ff1b6ba7544be046daf4ff22f5361fa04206`; all127 mappings were compared without remapping/rewrite.

## Initial v1.2 validation (historical)

Actual target MySQL8.4.8 accepted all migrations and views. Running Go with the restricted user passed loopback HTTP tests for PVP two-report confirmation/idempotency, PVE recording, local/BC single-reporter recording, normal abort merge, conflicting outcomes/wins, and successful matching completion. Eight QA matches were stored with qa-* versions; integrity view reportedzero errors. A Go-only restart retained all8 rows and restored health. Evidence: production-loopback-smoke.json, production-local-pvp-smoke.json, production-abort-smoke.json, production-pre-tls-validation.txt. No gameplay data was deleted and no tests reset production tables.

External HTTPS acceptance passed with system certificate/hostname verification enabled: exact Origin https://dotapk.lol and OPTIONS succeed, foreign Origin rejected403 without ACAO; public anonymous PVE completion persisted and GET/replay matched. The same match was independently verified in MySQL. Evidence: production-public-https-smoke.json, production-tls-validation.txt, production-existing-services-check.txt. A first probe issued immediately after nginx reload briefly hit the old certificate; subsequent actual SNI and external TLS probes both verified the dedicated certificate, without bypassing TLS checks. The existing Certbot cron is unchanged; no additional timer/service was added.

Remaining product acceptance: real-browser frontend domain PVP/gameplay tests coordinated by the parent task. Backend HTTPS/API acceptance alone does not claim completion of those browser tests.

## Rollback and operations

For the current room-fix registration, previous binary is `/opt/dota-duel/releases/v1.3-f041e4c6-heros22/dueld` (SHA256 `f041e4c6dde83461fd356849f31df0ff5d2837a2a05f5230445d5c11daa23dd1`). Before any6b match or live room exists, first revert frontend, restore that symlink and restart only DOTA DUEL; an exact compare-and-swap can revert the22 list from `[27c,6b]` to `[27c]`. Once6b is used, stop new starts/rematches, drain rooms/matches and retain used metadata/all records. Never restore a dump over production.

For the historical initial22 registration, prior compatible legacy20/46 binary is `/opt/dota-duel/releases/v1.3-124a519a-ui-code/dueld` (SHA256 `124a519ad6dda3be9a03eb76f4cc3032a1fc48ea58c4281d4779a95eadf5d718`). Before any new27c match or live22 room exists, switch frontend back first, restore that symlink and restart only DOTA DUEL. The deployment rollback can remove only its22 members and single roster row, never matches/reports. After any new22 usage, retain a compatible v1.3 reader/writer and drain rooms before a downgrade; keep all used registry metadata and records. Never restore a dump over production.

For the historical October6 UI registration, the prior compatible46 binary is `/opt/dota-duel/releases/v1.3-e6a9df66/dueld` (SHA256 `e6a9df663eb4816741fb84c6718bdc31d686d8e00745730957f97e0e376397f6`). Before any new-runtime match exists, restore that symlink and restart only `dota-duel.service`; revert only the exact `[e63,7e]` version list with compare-and-swap. After new-runtime usage, first stop new7e starts/rematches, finish or correctly abort sessions and close invitations; retain a compatible v1.3 reader/writer until drained. Keep every match/report and never restore an old dump over production.

The original October3 downgrade procedure is historical: before expanded rooms are used, code rollback restores the symlink to `/opt/dota-duel/releases/v1.2-07e2845b/dueld` (SHA256 `07e2845b60e8ac734ab4244fba193424de7d6b35dd997137ad02fc7b5852a3fb`) and restarts only `dota-duel.service`. Keep Nginx routes, unit, credentials, grants, additive tables/views and every record. After46 usage, first restore old20 frontend, stop new46 starts/rematches, finish or correctly abort existing46 sessions, close invitations and confirm no live/pending46 before downgrade. v1.2 is not a metadata-preserving46 writer; retain compatible v1.3 for inspection/history. Never overwrite an online database with an old dump, drop evidence, or touch business data. Restore only to an isolated recovery instance for investigation/reconciliation.

No other service restarts, MySQL setting changes, firewall/security-group expansion, host access keys or OAuth scope changes were performed. Monitor new DB/binlog growth and coordinate backup coverage: current AgentSquared backup implementation has not yet been changed to include this new schema. PVP peer agreement and PVE/local reports remain noncompetitive anonymous statistics; filter QA versions and keep transport/trust cohorts separate.
