# DOTA DUEL Go/MySQL backend

Independent backend for six-digit room invitations, WebRTC signaling, anonymous sessions, immutable P2P report reconciliation and clearly labeled PVE, same-screen and BroadcastChannel PVP self-reports. Combat remains in browsers. Source does not modify the frontend or AgentSquared application. Production Go+isolated schema run on the existing host; https://api.dotapk.lol is live with strict HTTPS/CORS and persistent-results checks passed. See DEPLOYMENT.md for current state.

- [Frontend API contract](docs/API.md)
- [AgentSquared inspection, isolation and deferred approvals](docs/DEPLOYMENT.md)
- [Validation evidence](docs/TESTING.md)
- `migrations/001_init.sql`: MySQL8.4+ schema. `002_analytics.sql`: original strict cohorts. `003_local_pvp_analytics.sql`: additional v2 views separating version/mode/transport/trust/hero/opponent/AI-difficulty cohorts. `004_hero_registry.sql`: additive frozen identities, roster membership and v3 views.
- `deploy/dota-duel.service`: loopback systemd unit with resource restrictions.
- `Dockerfile`, `compose.yaml`: local-only optional test stack; prohibited for this AgentSquared deployment, which reuses its existing MySQL.

## Build/test

Go1.26 and MySQL8.4+ target. Dependency github.com/go-sql-driver/mysql1.8.1 and transitive edwards25519 1.1.0 are pinned with go.sum; local build reused existing verified module cache. No production credentials are needed to build.

```sh
go test -race ./...
go vet ./...
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o bin/dueld-linux-amd64 ./cmd/dueld
```

Unit tests use a test-only in-memory transactional store. **The running server has no memory fallback**: writes fail when MySQL is unavailable. Integration tests additionally exercise real SQL, row locking/retries, generated columns/views and housekeeping.

For a disposable Unix-socket MySQL instance whose path includes `duel-mysql-test-` or `.mysql-integration`, first create `dota_duel`, then set `DUEL_TEST_DSN` to that socket DSN and run `go test -race -v ./...`. The integration suite applies migrations and deletes all rows in its dedicated tables between cases; it refuses TCP or unrelated sockets. Never point tests at production. Do not run destructive integration tests concurrently with QA against the same temporary database.

## Runtime

Provide `DUEL_MYSQL_DSN_FILE` pointing at the dedicated secret (preferred); `DUEL_MYSQL_DSN` also works for temporary local tests. Only database name `dota_duel` is accepted. Optional `DUEL_LISTEN` defaults to `127.0.0.1:18082`, `DUEL_ALLOWED_ORIGIN` is a single exact frontend origin, `DUEL_TRUSTED_PROXY_IP` is one exact proxy address (default disabled). No credential/request-body logging. Health `/healthz` verifies MySQL. Maintenance every30 seconds recycles expired invitations, erases SDP, expires sessions/limits/idempotency entries and marks unreported matches aborted.

MySQL migrations run out-of-band; API runtime user needs DML only. Analytics views are intentionally not exposed as public unauthenticated endpoints; use an authorized SQL reporting session:

```sql
SELECT * FROM dota_duel.duel_hero_balance_v2 WHERE game_version = 'the-exact-build' ORDER BY mode,hero,opponent_hero;
SELECT * FROM dota_duel.duel_data_quality_v2 ORDER BY game_version,mode,status;
```

`confirmed` PVP = both clients agreed, not anti-cheat proof. `recorded` PVE/local/BC PVP = client self-report, separated by mode and transport. Aborted/disputed/pending games are excluded from normal win-rate cohorts. Hero indices are scoped to game version. AI seat is excluded from PVE player win rate, but opponent hero/difficulty are retained. Mean duration is server start-to-final-report receipt and includes transport/report delay. Very small cohorts should not drive balance decisions.


Production contract is `v1.3-gameplay-rosters` (2026-10-03), accepting only legacy20 and `arena-first22-46-v1` bound exactly to `duel-e63dafb5ae2070a90f8b`. The source default deliberately remains legacy20; production uses [the frozen deployment profile](deploy/production-v13/gameplay-rosters.json). The127 identities do not make every hero playable; no24 or later candidate is activated.

The running release is reproducible from source `aa04e01f25ef2226b213cfc64e68c6c96fd0e18c` plus that profile with Go1.26.1:

```sh
python3 scripts/build-production-v13.py
```

This uses the local pinned Git snapshot and existing module cache, checks the exact Linux binary SHA256, and never accesses production or MySQL. [Build metadata](deploy/production-v13/build-metadata.json), [deployment receipt](docs/production-v13-deployment.json), [HTTPS checks](docs/production-v13-https-evidence.json) and [SELECT-only SQL checks](docs/production-v13-sql-evidence.json) make the source/profile/migration/runtime traceable. Evidence is a point-in-time backend acceptance snapshot, not a claim of subsequent browser gameplay. This handoff includes no credentials, raw submissions or full database dumps.
