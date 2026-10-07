[English](README.md) | [简体中文](README.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# DOTA DUEL Go / MySQL backend

API for [dotapk.lol](https://dotapk.lol): anonymous sessions, six-digit invitations, WebRTC signaling, immutable report reconciliation and separated result statistics. No account login is required. Combat runs in browsers; Go does not execute hero skills.

| Repository | Responsibility |
| --- | --- |
| [frontend](https://github.com/dotapk-lol/frontend) | Cloudflare static client, browser world/host, UI/input/rendering/AI and P2P |
| backend (this repository) | `https://api.dotapk.lol/api/v1`, Go service and MySQL persistence |
| [heros](https://github.com/dotapk-lol/heros) | MIT rules/parameters/host contract; [manual aggregate snapshots](https://github.com/dotapk-lol/heros/blob/main/balance-data/README.md) |

Production reuses the existing AgentSquared host, MySQL8.4 and Nginx with a dedicated `dota_duel` schema. The API already uses MySQL; no database switch, extra database or automatic balance export is needed. Source visibility does not authorize production operations.

## Local start

Go1.26.0+ (go.mod), MySQL8.4+, optional Python3 for profile helpers. Dependencies are pinned in go.sum. With an existing isolated local schema, applied migrations and runtime DML account:

```sh
export DUEL_MYSQL_DSN_FILE='<ABSOLUTE_PATH_TO_LOCAL_DSN_FILE>'
export DUEL_LISTEN='127.0.0.1:18082'
export DUEL_ALLOWED_ORIGIN='http://127.0.0.1:4173'
go run ./cmd/dueld
```

DSN file format: `<LOCAL_DB_USER>:<LOCAL_DB_PASSWORD>@tcp(127.0.0.1:3306)/dota_duel`. Fill placeholders privately; never commit/print actual credentials. The file takes precedence over `DUEL_MYSQL_DSN`. `DUEL_LISTEN` defaults to127.0.0.1:18082; `DUEL_ALLOWED_ORIGIN` is one exact origin; `DUEL_TRUSTED_PROXY_IP` is one optional exact proxy address, disabled by default. Health GET `/healthz` pings MySQL. There is no in-memory server fallback.

Ordinary builds contain **legacy20 only**. Current frontend play uses **22 heroes/88 slots**, `arena-heros22-v1`, `duel-heroes-127-v1`, source build `duel-851e67d77307f479f1fa`. Follow [development](docs/DEVELOPMENT.md) for the22 overlay, SQL metadata and exact binding. A catalog ID does not imply playability; unreleased heroes remain paused/grey.

## Documentation and checks

- [Architecture / directory map](docs/ARCHITECTURE.md)
- [Development / CORS / build registration](docs/DEVELOPMENT.md)
- [API contract](docs/API.md)
- [Testing](docs/TESTING.md)
- [Deployment and pinned profile](docs/DEPLOYMENT.md), [profile details](deploy/production-v13/README.md)
- [Contributing](CONTRIBUTING.md)

Run commands serially:

```sh
GOMAXPROCS=2 go test -p=1 ./cmd/... ./internal/...
GOMAXPROCS=2 go vet -p=1 ./cmd/... ./internal/...
python3 scripts/test-production-v13.py
```

Unit/profile tests use a test-only memory store. MySQL integration needs a dedicated disposable allowlisted Unix socket and clears its tables; see testing, never point it at production or retained QA data. Full tests, builds and deployment are unnecessary for prose changes.

PVP/WebRTC `confirmed / peer_agreement` means two identical completed reports, not anti-cheat proof. PVE/local/BC `recorded / client_reported` are separate cohorts; BC host reports once. Aborted/disputed/incomplete or abnormal games do not enter win rates. Internal30-second cleanup is not balance-data synchronization; public snapshots require manual review and aggregation with no identifiers/raw reports. Analytics views have no public unauthenticated REST/export endpoint.

Project-owned code/docs are [MIT](LICENSE), copyright2026 dotapk-lol contributors. Third-party images, music, trademarks and dependencies keep separate licenses/notices outside this grant. No Valve endorsement is implied.
