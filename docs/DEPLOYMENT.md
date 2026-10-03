# DotaPK production deployment

Current backend state (2026-10-03): **Go v1.3 and additive migration004 are deployed on the existing AgentSquared host; https://api.dotapk.lol external HTTPS/CORS and old20/new46 API acceptance passed.** Frontend publication/DNS belong to the separate frontend task. These records capture backend handoff before frozen46 publication and do not claim later browser gameplay.

## Repository and runtime

- Private repository: https://github.com/dotapk-lol/backend
- Host: existing `a2-webserver`, `43.162.87.40`, x86_64, OpenCloudOS,2 CPUs, about6.2GB available memory/69GB free disk at inspection.
- Game service remains `dota-duel.service`, created during the authorized initial deployment and reused for this upgrade. No new MySQL instance, Docker, Redis, TURN, proxy daemon, certbot scheduler or other service was added.
- Binary `/opt/dota-duel/releases/v1.3-e6a9df66/dueld`; `/opt/dota-duel/bin/dueld` symlink. Linux SHA256 `e6a9df663eb4816741fb84c6718bdc31d686d8e00745730957f97e0e376397f6`. Source snapshot `aa04e01f25ef2226b213cfc64e68c6c96fd0e18c` plus `deploy/production-v13/gameplay-rosters.json`, Go1.26.1; `scripts/build-production-v13.py` reproduces and checks the exact binary locally.
- Production profile is only legacy20 plus `arena-first22-46-v1`, exact build `duel-e63dafb5ae2070a90f8b`. No catalog127/CORE4-24/later candidate activation. A generic default build still enables only legacy20.
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

## v1.3 production release acceptance

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

Before expanded rooms are used, code rollback restores the symlink to `/opt/dota-duel/releases/v1.2-07e2845b/dueld` (SHA256 `07e2845b60e8ac734ab4244fba193424de7d6b35dd997137ad02fc7b5852a3fb`) and restarts only `dota-duel.service`. Keep Nginx routes, unit, credentials, grants, additive tables/views and every record. After46 usage, first restore old20 frontend, stop new46 starts/rematches, finish or correctly abort existing46 sessions, close invitations and confirm no live/pending46 before downgrade. v1.2 is not a metadata-preserving46 writer; retain compatible v1.3 for inspection/history. Never overwrite an online database with an old dump, drop evidence, or touch business data. Restore only to an isolated recovery instance for investigation/reconciliation.

No other service restarts, MySQL setting changes, firewall/security-group expansion, host access keys or OAuth scope changes were performed. Monitor new DB/binlog growth and coordinate backup coverage: current AgentSquared backup implementation has not yet been changed to include this new schema. PVP peer agreement and PVE/local reports remain noncompetitive anonymous statistics; filter QA versions and keep transport/trust cohorts separate.
