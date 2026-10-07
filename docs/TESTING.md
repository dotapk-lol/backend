[English](TESTING.md) | [简体中文](TESTING.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# Testing

Run commands serially with bounded concurrency. Normal unit tests exercise memory transactions, strict HTTP/JSON, membership/build validation, immutable retries, reconciliation, timeout/abort, local/BC trust and historical wire compatibility. They are not browser/WebRTC/native skill certification.

```sh
GOMAXPROCS=2 go test -p=1 ./cmd/... ./internal/...
GOMAXPROCS=2 go vet -p=1 ./cmd/... ./internal/...
python3 scripts/test-production-v13.py
```

Optional `GOMAXPROCS=2 go test -race -p=1 ./cmd/... ./internal/...` needs the platform race runtime. Do not set DUEL_TEST_DSN for memory tests. The profile script uses GOPROXY=off and a temporary Go overlay to check the retained production roster with memory tests; verified dependencies must already be available in the local Go cache. It creates no database or service.

Target `cmd` and `internal` explicitly: `deploy/production-v13/heros22_profile_test.go` is an overlay fixture for package `duel`, not a standalone Go package. A repository-wide `./...` includes that fixture directory and fails to compile; the profile script supplies the required overlay.

## MySQL integration

Existing tests in `internal/duel/mysql_test.go` require `DUEL_TEST_DSN` for a **dedicated disposable Unix socket** whose path contains `duel-mysql-test-` or `.mysql-integration`. Only schema `dota_duel` is accepted. Tests reject TCP/unrelated or retained QA sockets, apply migrations and delete dedicated tables between cases. The developer must provision that disposable environment separately; do not use production or a database containing records to preserve.

SQL checks cover real transactions/locking/retries, views/generated fields, identity membership and maintenance, with the actual driver. Integration may be skipped when the explicit DSN is absent; report that honestly. Never run destructive integration concurrently with another QA session against the same database. Do not publish DSN, token, player/match IDs, raw submissions or per-game receipts; keep local evidence ignored and submit only coarse reviewable summaries.

## Read-only local probes and documentation

Use the development guide's local health/registry/OPTIONS probes for CORS and profile visibility. They do not validate a natural match. Real browser/device, cross-network/NAT and persistence acceptance require separate controlled fixtures and explicit scope; no TURN or anti-cheat guarantee is implied.

For prose/template changes check Markdown pairing, language/website headers, relative links, script paths and `git diff --check`. Pinned binary reproduction is a build, not an ordinary documentation check. Historical logs removed from current source remain reachable in Git history; they cannot serve as current validation.

Paired-guide maintenance: `python3 scripts/check-docs.py` checks all retained Markdown headers, language partners and relative links.

Current source adds exact-build-gated [room-first selection and same-room rematches](SELECTION.md). Default feature bindings are empty; this is not a production activation. Memory/race and disposable MySQL tests cover epoch replay, own-seat locks, immutable prior matches, simultaneous locks, single allocation, expiry and legacy request bytes.
