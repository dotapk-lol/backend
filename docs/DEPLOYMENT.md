[English](DEPLOYMENT.md) | [简体中文](DEPLOYMENT.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# Deployment and operational boundaries

The project reuses existing AgentSquared MySQL8.4/Nginx with a dedicated `dota_duel` schema. Go listens on127.0.0.1:18082, and `https://api.dotapk.lol` routes through Nginx; Cloudflare hosts `https://dotapk.lol`. No new database/server, automatic balance sync or public statistics endpoint is part of source maintenance.

Retained operator inputs: `deploy/api.dotapk.lol.conf`, `deploy/dota-duel.service`, `deploy/public.env`, `deploy/provision-database.py`, `deploy/check-runtime-access.py`, and the [production-v13 profile](../deploy/production-v13/README.md). Public.env contains nonsecret runtime settings; the DSN is separate at an operator-controlled private path. Provision/check scripts are tied to this application's isolation policy, not portable setup commands; do not execute them as a contributor or fabricate replacement credentials. Existing public-domain and loopback routing values are intentional functional settings.

For a separate installation, review your own OS/service user, binary/secret/certificate paths, schema grants and exact origin; document examples with `<HOST>`, `<LOCAL_BINARY_PATH>`, `<DSN_FILE>` rather than real private hosts/IPs/user paths. Never substitute placeholders into a running production configuration blindly. The optional compose stack is for isolated local/staging environments and does not replace existing services.

## Profile, migration and release order

Ordinary source embed contains legacy20. Production profile preserves legacy20, historical46 and current22 exact build bindings; the frontend enables only22/88. Apply reviewed additive migrations/roster SQL through the administration role, build the matching embed/overlay, verify health/registry/CORS and then release the matching frontend. Runtime needs DML only. SQL INSERTs and compare-and-swap append scripts are not unconditional repeatable provisioning. See [development](DEVELOPMENT.md) and [API](API.md).

`python3 scripts/build-production-v13.py` reproduces a pinned Linux artifact with exact Go1.26.1 and cached dependencies, without database/server access. It checks frozen source bytes/profile SHA and expected binary SHA. This historical reproduction proves that artifact, not current uptime or browser acceptance. [Profile README](../deploy/production-v13/README.md) lists its immutable inputs; do not weaken checks or repin merely to hide a mismatch.

Plan rollback around compatible records/rooms: drain roster-aware clients/sessions and preserve a compatible reader for existing expanded records. Retain additive tables/views and original reports; do not rewrite old results or reuse a build ID for changed rules. Operator-managed artifacts/backups belong outside tracked source. An ordinary cleanup/license/prose commit needs no deployment.

Detailed historical deployment receipts, production host/IP notes and per-game QA data were removed from current source. They remain reachable in Git history; this was normal file cleanup, not secret revocation or historical erasure. No actual credential was found in the lightweight review; report any later finding privately by type/path without printing values.
