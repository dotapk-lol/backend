# Validation evidence

## Production v1.3 — 2026-10-03 UTC

Target MySQL8.4.8 migrations1–4 and frozen20+46 Go release are deployed.49 external TLS/CORS/API checks passed; original22 payload hashes/29 report digests preserved,7 explicitly synthetic HTTP games/10 reports independently verified, all v3 integrity counters0. New46 HTTP smoke games are aborted; the one synthetic completion is isolated under `qa-v13-legacy20`. This does not claim natural46 browser acceptance. See `production-v13-deployment.json`, `production-v13-https-evidence.json` and `production-v13-sql-evidence.json`. The exact local-only binary reproduction check is `production-v13-build-reproduction.json`.

All following sections are historical local/earlier deployment evidence from2026-10-02. Statements about pending production work reflect that stage, not the current production status above. Local QA instances remain separate and have not been reset by this release.

# Historical validation — 2026-10-02 (Asia/Shanghai)

## v1.3 registry candidate — local validation only

### Independent46 candidate service (18084, retained for browser QA)

Frozen frontend commit `3103f15607fd21e2817d7db330f4bd84a048ce04`; actual candidate `dist-candidate/client/build-manifest.json` and modules agree on `duel-e63dafb5ae2070a90f8b`, `arena-first22-46-v1`, and all46 explicit hero IDs. The stable source profile intentionally has another runtime and24 heroes; verification imports the built candidate modules. Every candidate identity matches the unchanged127-hero frozen registry. The roster is old24 plus21/28/29/32/36/42/47/50/55/57/58/62/71/81/82/94/99/104/106/117/121/124, preserving the manifest's exact IDs and order.

API `http://127.0.0.1:18084/api/v1`; exact CORS `http://127.0.0.1:4185`. The new MySQL instance uses `/tmp/duel-qa-candidate46-20261002/mysql.sock` with TCP disabled, independent data and a128MiB buffer pool. It is not a persistent system service. Candidate process metadata, overlay, local-only registration SQL, private0600 backup and raw verification receipts are under `../qa-runtime/candidate46/`. Resume with `python3 scripts/start-candidate46-qa.py`; it refuses to replace a listener or reset an existing data directory. The task-owned initialization requires explicit `--initialize` and nonexistent data. Its socket name is outside the destructive integration test allowlist.

Only this candidate binary includes46 through a build overlay. The repository's default embedded roster remains legacy20; production configuration/migrations are unchanged. The new instance also retains legacy20 and the previous24 roster/bindings as separate metadata. The46 roster binds only `duel-e63dafb5ae2070a90f8b`, with no wildcard. Original18083 and18082 were not restarted; old18083's22 full match payload hashes compare exactly equal before and after setup. Health probes confirm18082 remains v1.2 and both candidate services v1.3.

Seven live checks passed using the actual candidate `MatchAPI`: registry verified, all46 member creation/cancellation, PVE/local/BC recorded semantics, wrong roster/build and inactive-ID rejection, room/join checks plus immutable two-party confirmation, old omitted-roster wire/idempotency compatibility, and exact4185 CORS. Independently verified51 synthetic matches,52 original report bodies/digests and two participant mappings per match in SQL. Completed cohorts have7 appearances/4 wins across distinct trust/transport dimensions;47 cancelled membership/legacy probes are excluded. No mapping loss, join fanout or quality issues. Evidence: [HTTP](candidate46-http-evidence.json), [SQL, backup and preservation](candidate46-sql-evidence.json).

These are synthetic API/SDP tests. Their known IDs remain in the new test database for audit; they must not be described as natural browser outcomes. Browser rendering, skill semantics and cross-device acceptance remain pending. This service preparation does not enable46 on production.

### CORE4 local candidate activation (current18083 configuration)

Frontend commit `3fad3f881e38b83952cf44ceefbe9728f1ef6dff`, runtime `duel-751bcab20194934a863a`, roster `arena-core4-24-v1`, IDs0–19 plus25/31/45/100. On explicit local-test authorization, registered only this exact build on the independent18083 database. Existing legacy roster remains available to old builds. This is candidate test enablement, not production registration or skill/browser acceptance.

The repository's embedded default `registry/gameplay-rosters.json` still contains only legacy20. Candidate binary `bin/dueld-v13-core4-local` was built using `../qa-runtime/registry-b0/core4-build-overlay.json`, replacing that embed input only for this build with `core4-gameplay-rosters.json`. Local-only registration SQL, private0600 schema backup metadata and prior-row hashes are retained in the same runtime directory. The schema backup disables global GTID restoration statements. No production configuration or migration file was modified to activate this roster.

Only Go18083 was restarted (oldPID66388 → newPID79657); MySQL PID66345 was not restarted. All five preexisting match payload hashes are identical both immediately after registration and after smoke tests. The startup helper now resumes the binary saved in its process metadata; an explicit `--binary` accepts only this repository's bin directory and never replaces an existing listener. Original18082 remains v1.2.

Nine real HTTP checks using the actual CORE4 MatchAPI modules passed. Eight synthetic matches were independently verified in SQL. Four appended IDs map25→Valve15,31→28,45→47,100→102; each match produces exactly two resolved participant rows, with PVE counting only the human seat. Test-scoped cohorts total10 CORE4 appearances plus1 legacy appearance, without fanout or dropped mappings. Tests cover create/join rejection of unknown roster/build, omitted or legacy fallback for the bound new build, inactive hero rejection, unchanged result digests, legacy omitted-body creation/join/idempotency, local/BC single-reporter semantics and two-peer confirmation. Evidence: [HTTP](core4-local-http-evidence.json), [SQL and preservation](core4-local-sql-evidence.json).

The service is ready at `http://127.0.0.1:18083/api/v1`, with exact CORS `http://127.0.0.1:4174`. A page opened before the switch may retain a rejected `MatchAPI.registryTask`; reload normally to create a new adapter and fetch the current roster. Do not weaken registry validation. Any later gameplay build needs a separately authorized exact candidate binding; no wildcard was enabled.

### B0 live compatibility service (historical legacy-only checkpoint)

API `http://127.0.0.1:18083/api/v1`, health `http://127.0.0.1:18083/healthz`; exact CORS origin `http://127.0.0.1:4174`. `localhost:4174`, the old4173 origin and the production origin are not allowed on this candidate. Production CORS has not changed. Frontend B0's current default local base is still18082: its single writer must explicitly select18083 for this isolated preview; opening4174 alone does not select this API.

Go binary `bin/dueld-v13-registry-b0`, isolated MySQL socket `/tmp/duel-mysql-test-registry-b0-20261002/mysql.sock`, data under its own `data/`, TCP disabled. Process metadata and logs live outside the repository at `../qa-runtime/registry-b0/`. Resume with `python3 scripts/start-registry-b0-qa.py`; it reuses existing data and refuses to replace a listener. `--initialize` is for the first nonexistent B0 data directory only, never a reset. Do not point destructive integration tests at this retained database; the test guard explicitly rejects its socket as well as the original18082 database socket.

Imported the actual B0 `MatchAPI`/`resultPayload` modules read-only from frontend commit `8ad8bc5534b3ff0580c9328d4a62028f0eb74726`, runtime `duel-12e243d1c6a8a142b361`. Nine live HTTP checks passed: catalog verification, explicit roster creation/replay, changed-body409, old omitted-roster wire format, local/BC adapters, locked catalog heroes/proposed subset rejection, mixed omitted/explicit PVP clients with version409 and two-report confirmation, exact CORS, and preserved18082 v1.2 health. SELECT-only verification checked five persisted matches, unchanged result digests and the original legacy creation digest. Evidence: [HTTP](v13-b0-http-evidence.json), [SQL](v13-b0-sql-evidence.json). These are synthetic outcomes and valid placeholder SDP, not real browser WebRTC gameplay.

Only `legacy-20-v1` is enabled. Proposed `arena-core4-24-v1` (0–19 plus25/31/45/100) is not registered and has no accepted build. Browser acceptance and later selected-kit implementation remain with the frontend task. No18082 restart, original QA data changes, production deployment or production CORS change occurred.

- Frozen mapping verified against all20 historical hero names, indexes and Valve IDs. Catalog hash `5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138` matches canonical heroes JSON; original manifest copied verbatim. 127 known identities, only `legacy-20-v1` gameplay enabled.
- Full `go test -race -v ./...` passed105 test/subtest cases, including real MySQL9.6 over the new `/tmp/duel-mysql-test-registry-20261002/mysql.sock`. TCP disabled. `go vet ./...` and static Linux amd64 build passed. Complete evidence: [v13-registry-mysql-tests.txt](v13-registry-mysql-tests.txt), [v13-registry-evidence.json](v13-registry-evidence.json).
- Verified original room/PVE/local/create-match/Result JSON byte sequences and digests; replayed a persisted v1.2 room and request record without metadata. Unknown/disabled heroes and invalid roster requests rejected across all four entrypoints. Fixture-only expanded gameplay subset enforces exact builds, inheritance on rematch, two-party confirmed and single-party recorded semantics. HTTP tests preserve required/null/unknown field validation and expose only public match metadata.
- Applied migration004 twice on the independent database, verifying old full-payload hashes unchanged, v1/v2/v3 historical counts equal, no join fanout, and no dropped unmapped records. All127 SQL identity mappings match the manifest. Actual fixture subset IDs20/126 map to Valve3/155; fixture acceptance was removed afterward and is absent from the shipped manifest/migration. PVP/PVE/local/BC cohorts remain separated and pending/aborted excluded.
- Read-only health confirmed the existing18082 QA service still reports `v1.2-abort-reconciliation`; neither its database nor its process was restarted/reset. Integration tests now explicitly reject the preserved QA socket even if mistakenly configured. The new registry test MySQL instance was shut down after testing.
- Candidate binary: `bin/dueld-v13-registry-linux-amd64`, SHA256 `c04b30b17e06fb2a6d507c796c74b1ea80d2a9c920f96935a7398691e289b244`. This is not a deployment or frontend gameplay acceptance. Production remains v1.2. Target MySQL8.4 migration validation, accepted expanded gameplay roster/build bindings, coordinated frontend rollout and production release remain future work.

Historical local validation below predates production deployment. Current production Go/MySQL loopback acceptance and remaining DNS/TLS boundary are documented in DEPLOYMENT.md and production-*-smoke evidence; no AgentSquared business data was modified.

## Passed

- Go1.26.1 `go test -race -v ./...` against in-memory transactional test store and a real, isolated MySQL9.6.0 Unix socket: PASS.
- 12 domain scenarios each ran against both stores: leading zeros/collision retry/code recycling, 12 concurrent guests competing for one seat, room idempotency/self-join/version/membership, two-party confirmation and concurrent immutable retries/rematch IDs, disputes/outsiders, single-party timeout/late report, no-report timeout, PVE without AI credentials, aborted games never awarding wins, illegal round/score/winner/version structures, 30 parallel rate-limit requests (exactly10 admitted), session expiry.
- Additional SQL test: migrations1/2, persisted timeout maintenance, expired SDP scrub, rematch after SDP expiry, PVP aggregate includes2 player appearances, PVE includes only1 human appearance, pending/aborted excluded, integrity view reports0 invalid results.
- HTTP decoder/CORS tests: valid request, rejected foreign origin, missing/null/unknown/oversize/trailing data; invalid fixed-array score lengths; peer reports hidden; random code format; rollback behavior. Race detector reported no races.
- `go vet ./...`: PASS.
- Native macOS arm64 and static Linux amd64 binaries built. Current v1.2 target Linux SHA256: `07e2845b60e8ac734ab4244fba193424de7d6b35dd997137ad02fc7b5852a3fb`.
- Actual HTTP -> running Go -> real MySQL smoke: health, two distinct anonymous sessions, six-digit room creation, guest claim/answer, both ready, first report pending, second confirmed, duplicate immutable report successful, malformed score rejected, PVE recorded/client_reported. Evidence: [http-smoke-evidence.json](http-smoke-evidence.json). No tokens stored in evidence.

## Local QA service retained for frontend integration

`http://127.0.0.1:18082`, exact CORS origin `http://127.0.0.1:4173`. API executable `bin/dueld-local`, MySQL Unix socket `/tmp/duel-mysql-test-20261002/mysql.sock`, no TCP DB listener. The temporary database has no business data and is separate from any existing local MySQL instance. Its initialization used an ephemeral local test root with no password, accessible through the temporary Unix socket only; it must never be used as a production configuration. No persistent server credential was created.

Do not rerun `TestMySQLIntegration` while frontend QA is using this instance: that suite deletes its isolated test tables between cases. It is safe to run `go test -race -run 'TestMemorySuite|TestHTTPBoundaries|TestNoResultLeak|TestRandomCodes|TestStoreRollback' ./...` during QA. After coordinated QA completion, terminate the Go and temporary MySQL processes and remove only this task's temporary data directories; do not stop existing MySQL services. Local process lifetime is not production availability.

## Limits / still required

- Production MySQL is8.4.8; local real-server validation used9.6.0. SQL is written for8.4+, but target-version migration/smoke remains required in the new isolated schema after authorization. No production access was used for test writes.
- Docker is unavailable here and on selected host; Docker/Compose templates were prepared, not executed. systemd unit prepared, not installed.
- Frontend integration and real-browser PVP/PVE persistence checks belong to the frontend single writer/parent task. This backend-only smoke uses valid synthetic SDP/results, not live WebRTC combat.
- No public HTTPS route, domain, new MySQL runtime credential, production backup integration or deployment has been enabled. Those are explicitly deferred.
- Anonymous statistics and peer agreement do not establish human identity or eliminate collusion/client tampering. PVE is self-reported. Duration is server timestamp/receipt duration, not exact combat-frame duration.

## Additive v1.1 local/BC PVP validation

- Added `/matches/local` without changing existing P2P/PVE input contracts; `/healthz` reports `contractVersion: v1.1-local-pvp`.
- Second isolated MySQL socket `/tmp/duel-mysql-test-local-pvp-20261002/mysql.sock` used for the entire v1 suite plus local-PVP tests. Original QA database was not cleared. Full real-MySQL/race log: `v2-mysql-test-evidence.txt` (PASS).
- Local and BroadcastChannel scenarios each verified two unique match-local anonymous slots, single real reporter, 12 parallel idempotent create retries, outsider rejection, altered payload conflict, immutable single-party recorded result, and rematch ID/slot separation. Abort/timeout, invalid transport/hero, missing fields and attempted trust promotion also tested.
- Migration003 adds v2 balance/data-quality views only. Tests confirm old view excludes local PVP, v2 has separately filtered local and BC cohorts (2 appearances per completed game), peer-confirmed cohort remains2, and no invalid trust promotions.
- Real HTTP on candidate18083 passed (`v2-local-http-evidence.json`). With the parent’s coordinated approval, applied only additive migration003 to original QA database, stopped only old Go PID12855, and started new Go on18082. No MySQL restart or deletion. HTTP on upgraded18082 passed (`v2-upgraded-18082-evidence.json`).
- Independently verified frontend natural PVE0:2 match `c6ad274c418c44c79bd0afd2bc5e31ae5d003a63d660cc3a50c9865ce02cee97`, build `duel-87a4da42c0a471d515e7`, recorded/client_reported, normal difficulty, duration58.017s. Full selected-row JSON and payload SHA256 before/after upgrade compare exactly equal (`pve-before-upgrade.json`, `pve-after-upgrade.json`).
- At the v1.1 handoff,18082 served the v1.1 binary connected to original QA MySQL; old records retained. Production still untouched. Actual frontend local/BC integration QA remains with the parent/single writer.

## v1.2 abort-reconciliation fix

Browser QA found two normal departures marked disputed because aborted reports had different `left`/`disconnect` reasons. Both original records were independently verified in SQL; the fix changes future reconciliation and leaves historical evidence unchanged.

- Both valid aborted reports now produce aborted/no winner, preserving all report bodies/digests/receipts. Same reason retained; different reason summarized as interrupted. Equal scores receive peer_agreed metadata; different partial scores remain unresolved. Completed-vs-aborted stays disputed/outcome_conflict. Completed-result disagreements remain disputed; exact completed agreement remains confirmed.
- 17-case matrix per store: 8 scenarios in both arrival orders plus timeout/late report. Covers pending replay deadline stability, terminal replay immutability, original report preservation, result conflicts and aborted/disputed exclusion from SQL win-rate views. Entire real-MySQL/race suite and go vet passed; log `v12-mysql-test-evidence.txt`. Tests ran only on the separate auxiliary database, not the preserved browser QA database.
- Upgraded only the Go process on18082 after coordinated QA release; no migration needed, no MySQL restart or data clearing. `/healthz` now reports `v1.2-abort-reconciliation`. HTTP evidence `v12-http-evidence.json`; independently read SQL receipts `v12-sql-receipts.tsv` show left/disconnect reports remain intact under aborted/interrupted, mixed outcomes and genuine winner disagreements stay disputed, identical completion confirmed.
- Independently checked all6 handback IDs: PVE0:2/local2:0/BC2:0 recorded; WebRTC2:0 confirmed with two reports; fresh rematch and prior disconnect are historical disputed rows with two aborted reports. `v12-handback-before.ndjson` and `v12-handback-after.ndjson` contain selected fields, original reports and full-payload hashes and compare byte-for-byte equal. No historical row was silently reclassified.
- Main Go18082 + original QA MySQL remain running. Auxiliary test MySQL was shut down after verification. Production unchanged. A real-browser disconnect recheck is still owned by the parent QA task; this backend evidence does not claim that browser rerun.
