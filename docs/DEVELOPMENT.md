[English](DEVELOPMENT.md) | [简体中文](DEVELOPMENT.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# Local development and22 integration

Use an existing developer-owned isolated MySQL `dota_duel`; never operate production from these examples. The server accepts only that schema name. Runtime accounts need business DML; a separate local administration session applies migrations/metadata. Keep DSNs, backups and raw reports out of Git.

## Migrations and roster metadata

Apply migrations001–004 in order. The following assumes the existing local schema has not registered the22 roster. `--login-path=<LOCAL_ADMIN_PROFILE>` is a privately configured MySQL CLI profile, not a shared credential:

```sh
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < migrations/001_init.sql
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < migrations/002_analytics.sql
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < migrations/003_local_pvp_analytics.sql
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < migrations/004_hero_registry.sql
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < deploy/production-v13/register-heros22.sql
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < deploy/production-v13/append-room-fix-runtime.sql
mysql --login-path='<LOCAL_ADMIN_PROFILE>' dota_duel < deploy/production-v13/append-i18n-runtime.sql
```

These are existing files, not new export/registration tools. Inspect the schema first: register-heros22 is an INSERT, not repeatable registration; the append scripts compare exact prior lists. Expect1 roster+22 members, then1 updated row per append;0 means a mismatched precondition, not success. Existing databases need reviewed missing steps, not wholesale replay. This guide does not automatically execute commands or create a database.

## Go embed profile

Default `go run ./cmd/dueld` / `go build` embeds only legacy20 from `internal/duel/registry/gameplay-rosters.json`; environment variables cannot enable another roster. The existing `deploy/production-v13/gameplay-rosters.json` preserves the full historical profile. Go overlay replaces that input without editing source:

```sh
python3 - <<'PYCODE'
import json, pathlib, tempfile
root = pathlib.Path.cwd()
p = pathlib.Path(tempfile.gettempdir()) / 'dotapk-local-roster-overlay.json'
p.write_text(json.dumps({'Replace': {
    str(root / 'internal/duel/registry/gameplay-rosters.json'):
    str(root / 'deploy/production-v13/gameplay-rosters.json')
}}))
print(p)
PYCODE
export DUEL_MYSQL_DSN_FILE='<ABSOLUTE_PATH_TO_LOCAL_DSN_FILE>'
export DUEL_LISTEN='127.0.0.1:18082'
export DUEL_ALLOWED_ORIGIN='http://127.0.0.1:4173'
GOMAXPROCS=2 go run -p=1 -overlay="${TMPDIR:-/tmp}/dotapk-local-roster-overlay.json" ./cmd/dueld
```

Use the exact Python output path if tempfile differs from shell TMPDIR. To build a binary use `GOMAXPROCS=2 go build -p=1 -overlay='<ABSOLUTE_OVERLAY_PATH>' -o '<LOCAL_BINARY_PATH>' ./cmd/dueld`; do not replace a running service. The temporary overlay has no credentials/data and may be deleted after stopping the process.

The profile retains legacy20/46 compatibility; it does not enable46 in the frontend, which still admits22/88. SQL metadata must match embed metadata to keep persistent/analytics identities complete.

## Exact bindings

| Field | Current22 profile |
| --- | --- |
| rosterId | `arena-heros22-v1` |
| registryVersion | `duel-heroes-127-v1` |
| registrySha256 | `5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138` |
| gameVersions | `duel-27c78aa4cfc8facc8a23`, `duel-6b1d12f75aa4bbac4e12`, `duel-851e67d77307f479f1fa` |
| Source frontend NET_VERSION | `duel-851e67d77307f479f1fa` |
| heroIds | `1,3,4,5,7,8,9,15,17,18,28,31,32,36,50,55,57,58,62,71,81,82` |

registryNumericId differs from Valve ID/array order. GET `/api/v1/registry` exposes identities/gameplayRosters; catalog presence is not playability. Expanded rosters cannot use wildcards. Current build with omitted rosterId gives409, excluded hero400; both peers match version/roster. Frontend checks all identities/hash; see [frontend development](https://github.com/dotapk-lol/frontend/blob/main/docs/DEVELOPMENT.md).

Rebuilding frontend may generate a version absent from the allowlist. Review/register the same exact version in a separate local profile copy and local SQL, then run that backend. Do not fake NET_VERSION, use catalog127 as rosterId or add a production wildcard. Docs do not rebuild game identity. Go does not execute heros rules; public88 and frontend composition hashes differ, see [compatibility](https://github.com/dotapk-lol/heros/blob/main/docs/release-compatibility.md).

## CORS and minimal probes

Frontend origin is exactly `http://127.0.0.1:4173`, matching DUEL_ALLOWED_ORIGIN. localhost differs; omit trailing slash/path. Direct local connections need no DUEL_TRUSTED_PROXY_IP.

```sh
curl --fail http://127.0.0.1:18082/healthz
curl --fail -H 'Origin: http://127.0.0.1:4173' http://127.0.0.1:18082/api/v1/registry
curl --fail -X OPTIONS -H 'Origin: http://127.0.0.1:4173' -H 'Access-Control-Request-Method: POST' -H 'Access-Control-Request-Headers: Content-Type, Authorization' http://127.0.0.1:18082/api/v1/sessions
```

These inspect local health/registry/preflight without generating reports. Verify roster/version before play, then use anonymous authenticated request shapes from [API](API.md). No WebSocket signaling/public analytics REST exists. WebRTC has no TURN; real NAT/device acceptance remains separate from source tests.

## Tests and analytics

README unit tests need no database. `scripts/test-production-v13.py` checks profile/build rejection using memory and a temporary overlay, serially. MySQL integration clears tables only on explicitly allowlisted disposable sockets; see [testing](TESTING.md).

`duel_hero_balance_v3` groups by build, roster/registry, mode/status/trust/transport, AI difficulty, seat, hero/opponent. PVP contributes up to two appearances per game, PVE only human seat0; appearances are not unique matches. Quality views do not automatically fix anomalies; review duplicate/time/identity/trust/test-game filters. Public data follows [manual balance policy](https://github.com/dotapk-lol/heros/blob/main/balance-data/README.md), never raw SELECT* output, match IDs or report bodies.

## License

Project-owned code/docs are [MIT](../LICENSE), with retained notices. Third-party media/fonts/trademarks/dependencies keep separate terms; preserve upstream LICENSE/NOTICE and review redistribution rights independently.
