# Validation evidence — 2026-10-02 (Asia/Shanghai)

Status: independent implementation validated locally; **AgentSquared production not deployed**. No production database, account, listener, nginx configuration or business data was modified.

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
