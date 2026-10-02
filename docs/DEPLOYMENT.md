# DotaPK production deployment

Current state (2026-10-02): private repository pushed; Go v1.2 and isolated MySQL schema deployed on AgentSquared; loopback acceptance passed. **Public API TLS/domain activation waits for api.dotapk.lol DNS.** Frontend/DNS are owned by the separate frontend task; this backend task does not modify Cloudflare DNS.

## Repository and runtime

- Private repository: https://github.com/dotapk-lol/backend
- Host: existing `a2-webserver`, `43.162.87.40`, x86_64, OpenCloudOS,2 CPUs, about6.2GB available memory/69GB free disk at inspection.
- Only new persistent service: `dota-duel.service`. No MySQL instance, Docker, Redis, TURN, proxy daemon, certbot scheduler or other service was added.
- Binary `/opt/dota-duel/releases/v1.2-07e2845b/dueld`; `/opt/dota-duel/bin/dueld` symlink. Linux SHA256 `07e2845b60e8ac734ab4244fba193424de7d6b35dd997137ad02fc7b5852a3fb`.
- Go binds `127.0.0.1:18082` only. systemd DynamicUser, MemoryMax128MiB, CPUQuota25%, TasksMax32; DB pool5 open/2 idle. User authorized the server deployment and dedicated identity.
- Exact CORS origin `https://dotapk.lol`; only trusted proxy `127.0.0.1`. Nginx overwrites X-Real-IP. Keep API DNS-only unless Cloudflare proxy client-IP trust is separately configured; never blindly trust incoming CF/X-Forwarded-For headers.

## Database isolation

Existing MySQL8.4.8 at127.0.0.1:3306 is reused. Dedicated schema `dota_duel`, migrations1/2/3 applied and tested. Existing business schemas and users were not modified. Schema separation is database permission isolation, not separate hardware/process or protection from the host administrator.

Inspection showed `agentsquared@127.0.0.1` has ALL PRIVILEGES on agentsquared_website, so it was not reused. With explicit user approval, created only `duel_app@127.0.0.1` with SELECT, INSERT, UPDATE, DELETE on dota_duel.*. Runtime login/read passed; reads of mysql.user and the AgentSquared business table were independently denied. No global, DDL, grant or other-schema privileges.

A32-byte CSPRNG secret was generated only on the server, never displayed or committed. `/etc/dota-duel/mysql-dsn` is root:root0600 under0700 directory; systemd LoadCredential exposes it to Go. Public non-secret origin settings are separate. Existing root administrator authentication was used only for authorized schema/provisioning operations; the game does not receive root or A2 business credentials. `deploy/provision-database.py` refuses existing identity/secret overwrite. Do not rerun it casually. `deploy/check-runtime-access.py` is a read-only verification helper.

## Nginx and certificate plan

Existing Nginx1.26.3, Certbot2.8 webroot `/var/www/certbot`, and existing `/etc/cron.d/agentsquared-certbot-renew` are reused. That existing cron renews daily at03:17 and reloads nginx through a deploy hook. No new renewal service/timer is required.

`api.dotapk.lol` requires its own certificate/key; AgentSquared/Pikoo certificates are not valid substitutes. HTTP-only challenge vhost is installed and nginx-tested/reloaded; root routing proof passed using an explicit Host. Public certificate issuance waits for DNS-only A=43.162.87.40, no AAAA. Once DNS actually resolves and the HTTP challenge is reachable, use existing ACME account to issue api.dotapk.lol certificate, then install the prepared443 vhost. Its only proxied routes are `/api/v1/` and `/healthz` to127.0.0.1:18082; no new public port is opened. Root frontend remains on Cloudflare.

Before changes, existing Nginx conf.d was copied to `/var/backups/dota-duel/initial-20261002/nginx-conf.d`; baseline hashes are in existing-nginx.sha256. A2/Pikoo files remain byte-for-byte unchanged after HTTP-stage reload. Existing A2 healthz returnedok; nginx/A2/MySQL/Redis stayedactive. A pre-existing unrelated tat_agent unit warning was observed, not modified.

## Validation performed

Actual target MySQL8.4.8 accepted all migrations and views. Running Go with the restricted user passed loopback HTTP tests for PVP two-report confirmation/idempotency, PVE recording, local/BC single-reporter recording, normal abort merge, conflicting outcomes/wins, and successful matching completion. Eight QA matches were stored with qa-* versions; integrity view reportedzero errors. A Go-only restart retained all8 rows and restored health. Evidence: production-loopback-smoke.json, production-local-pvp-smoke.json, production-abort-smoke.json, production-pre-tls-validation.txt. No gameplay data was deleted and no tests reset production tables.

Pending: API DNS/TLS and external HTTPS/CORS acceptance, then real-browser frontend domain battle/result tests coordinated by the parent task. The existence of a healthy loopback process does not imply public release completion.

## Rollback and operations

Stop/disable only dota-duel.service. Remove only the new api.dotapk.lol virtual host if its activation must be undone; nginx -t before reload. Restore only that new vhost's prior staged configuration if applicable, not all older service files. For code rollback, restore the previous DotaPK binary symlink and restart only this service. Keep dedicated database records and protected credentials; never automatically drop schema/user. Current initial deployment has no older DotaPK server release.

No other service restarts, MySQL setting changes, firewall/security-group expansion, host access keys or OAuth scope changes were performed. Monitor new DB/binlog growth and coordinate backup coverage: current AgentSquared backup implementation has not yet been changed to include this new schema. PVP peer agreement and PVE/local reports remain noncompetitive anonymous statistics; filter QA versions and keep transport/trust cohorts separate.
