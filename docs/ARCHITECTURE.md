[English](ARCHITECTURE.md) | [简体中文](ARCHITECTURE.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# Architecture and data grain

The HTTP handler authenticates anonymous sessions, validates strict bodies/origin/membership and applies persistent request limits. Service transactions own room seats, idempotency, readiness, deadlines and immutable report reconciliation. SQLStore persists to existing MySQL; only tests use a memory store. Browser combat is outside the server.

| Path | Role |
| --- | --- |
| `cmd/dueld/` | Environment, SQL ping, HTTP listener and30-second housekeeping |
| `internal/duel/http.go`, `service.go`, `local.go` | Routes, session/room/match lifecycle and local reporting |
| `internal/duel/reconcile.go`, `model.go` | Report validation, summary resolution and wire shapes |
| `internal/duel/store.go` | MySQL transactional persistence and cleanup |
| `internal/duel/registry/`, `registry.go` | Frozen identities, embedded gameplay memberships/builds |
| `migrations/001_init.sql`–`004_hero_registry.sql` | Persistence, v1/v2 analytics and roster/v3 metadata |
| `deploy/production-v13/` | Retained profile, reviewed additive registration SQL and profile test |
| `deploy/` | Existing operator Nginx/systemd/DSN-isolation setup, not general credentials |
| `scripts/build-production-v13.py`, `scripts/test-production-v13.py` | Pinned binary reproduction and memory profile checks |
| `Dockerfile`, `compose.yaml` | Optional isolated local/staging stack, not existing host deployment |

Server records snapshot game version/roster/registry. Match-local IDs for local/BC seats are not two authenticated people; one reporter owns their record. PVP reports are immutable per seat, raw submissions hidden from HTTP, and two reports agree only at the semantic reconciliation boundary. Cleanup expires invitations/SDP/sessions/limits and times out games, without exporting data.

Views group status/trust/transport as well as build, identity and opponent. A PVP game can yield two seat appearances; PVE analytics only counts human seat0. Raw match/report tables and views are internal; anonymous IDs remain sensitive. Check missing mappings, bad timestamps/winners/confirmation, duplicates and test games before aggregating; no automatic correction or public analytics endpoint exists. Aborted/disputed games do not count as wins.

Use [API](API.md) for exact input/output semantics and [manual policy](https://github.com/dotapk-lol/heros/blob/main/balance-data/README.md) for reviewed public snapshots. Removed historical receipts/logs remain in Git history; source cleanup does not erase server data or rewrite history.
