# AgentSquared isolation/deployment handoff

Verified 2026-10-02 Asia/Shanghai. No production mutations performed.

## Read-only evidence

Relevant local files: `/Users/didi/Project/AgentSquared/ENVIRONMENT.md`, deployment runbook, existing SSH alias configuration. Existing SSH key/known-host entry used with BatchMode and StrictHostKeyChecking=yes; no new key, forwarding, account, firewall, DNS or tunnel was installed. Do not copy the legacy runbook into this repo: it contains unrelated secrets.

- Selected candidate: `a2-webserver`, `43.162.87.40`, hostname `VM-0-8-opencloudos`, x86_64, 2 CPUs.
- ~7681 MB RAM total, 6244 MB available; root disk 80 GB, 69 GB available; load ~0.42/0.27/0.21 at inspection.
- nginx, agentsquared-webserver, mysqld, redis active.
- MySQL 8.4.8 loopback 127.0.0.1:3306; Redis loopback 6379. No public DB port.
- Existing API listeners: 8080 (AgentSquared), 127.0.0.1:18081 (Pikoo); public nginx 80/443; relay 4051; SSH22.
- Nginx `api.agentsquared.net` -> 127.0.0.1:8080 and `api.pikoo.lol` -> 127.0.0.1:18081. These are outside this task's modification scope.
- Proposed Go listener 127.0.0.1:18082 was unused; `/opt/dota-duel` did not exist.
- `a2-db` (43.162.125.53) is documented retired/read-only, not a target. `a2-hermes` is a separate test-agent machine, not needed. Existing topology unambiguously favors a2-webserver.
- Docker was not found on selected server; use independent systemd binary and existing MySQL with dedicated schema. No Docker installation planned.

## Proposed deployment, not executed

- Binary `/opt/dota-duel/bin/dueld`, separate systemd unit `dota-duel.service` using DynamicUser, CPUQuota=25% of one CPU, MemoryMax=128M, 32 tasks, 5 DB connections (2 idle).
- Only `127.0.0.1:18082`. No security group/firewall change. No public IP HTTP URL (existing HTTPS browser cannot safely call it anyway).
- Database `dota_duel` on the existing local MySQL. No reference to AgentSquared/Pikoo schemas. Application refuses a DSN with any other database name. Tables and analytic views are namespaced `duel_*`.
- Distinct schema/account isolates permissions/data, **not hardware failure, server capacity, root administration or MySQL process**. Existing MySQL has max_connections=100 and 512 MB buffer pool per runbook. Up to 5 added connections plus writes/binlog growth are incremental shared-resource risks. Need inspect size/latency after staging; do not increase MySQL settings automatically.
- No frames/replays/assets/music in MySQL. Transient SDP ≤40KB per description; erased after10min. Session/rate/request rows expire. Match evidence is retained for balance work; no automatic destruction policy yet. Monitor growth and coordinate backup retention before production.

## Exact deferred permission item

After user wakes, seek one concrete authorization covering the following (never reveal generated secrets):

1. Create **only** schema `dota_duel` on a2-webserver.
2. Create **only** runtime MySQL account `'duel_app'@'127.0.0.1'` (TCP loopback only) with `SELECT, INSERT, UPDATE, DELETE ON dota_duel.*`. No global privileges, GRANT OPTION, DDL, FILE, SUPER, administration, access to other schemas, or remote host wildcard. Migration runs using an already-authorized local administrator session; do not create a permanent migration account.
3. Generate a new 32-byte CSPRNG secret locally on the server, without echo/command-line history or placing secret in SQL/process arguments. Use a protected temporary administrator input file/pipe. Write the resulting DSN to `/etc/dota-duel/mysql-dsn`, owned root:root 0600 under a0700 directory. Load via systemd `LoadCredential`; never commit, copy into chat, add to build, or reuse existing business credentials. Remove any transient secret files immediately after successful installation. Do not enable MySQL general query logging. Account creation may be retained in administrator/binlog audit according to existing server policy.
4. Install only this new binary/unit/config and start the loopback listener after migrations/tests; no changes or restart to existing AgentSquared services or nginx routes.

The user has authorized the eventual isolated deployment but the explicit persistent-secret creation step remains deferred by the parent task's instruction. No permission question is being sent during sleep.

## Public release remains deferred

User will provide a new domain after waking. Do not register a domain, alter AgentSquared/Pikoo DNS/traffic, expose18082, or install a tunnel. Current private Sites URL is not yet wired to this host. Domain ownership/TLS/new nginx virtual host or a narrow Worker proxy must be planned with the actual domain, reviewed, then explicitly deployed. Exact allowed CORS origin and any trusted proxy IP must be set at that point. The proxy must overwrite X-Real-IP and preserve Bearer authorization; no caching. Keep old D1 signaling disabled when frontend switches to Go; never dual-write divergent results.

## Validation before enablement

Inspect dedicated account grants, table/view migration versions1/2/3, `/healthz`, two distinct sessions completing PVP, duplicate result replay, guest race, aborted game, PVE recorded game, analytics exclusions, restart persistence, MySQL8.4.8 behavior and existing service health. Use only test matches in this new database. Recheck listener conflict immediately before startup.

## Rollback

Stop/disable only `dota-duel.service`; retain database evidence and secret backup under existing secure handling. Restore previous DOTA frontend/API base if frontend was switched, and remove only a newly added DOTA virtual-host/proxy route if one was separately approved. Leave existing nginx configurations, AgentSquared services and all business schemas untouched. Revert binary via versioned copy/symlink after health verification. **Do not** automatically DROP DATABASE or DROP USER: those are destructive cleanup actions requiring deliberate authorization and a verified backup. Local staging Compose teardown also keeps the volume unless explicit data deletion is intended.
