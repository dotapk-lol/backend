[English](SELECTION.md) | [简体中文](SELECTION.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# Room-first selection and same-room rematches

This additive protocol is available in current source, but no new production frontend build is activated here. Exact approved builds opt in through `internal/duel/registry/protocol-features.json` (`roomSelectionVersions`); its shipped list is empty. Each version must already bind an accepted gameplay roster containing default hero1. Retain every previous roster/build binding when creating a reviewed overlay. New builds cannot bypass selection by omitting the endpoint or `selectionEpoch`. Existing builds preserve the original API and request digests; PVE/local/BC are unchanged.

Create/join reserve the existing seats with hero1, an independent session token and the unchanged six-digit invitation string. Establish and answer signaling, then use `POST /api/v1/rooms/{roomId}/selection` with the participant's Bearer token. Both participant sessions, exact version/roster, answered signaling and an open room are required. Connected room lifetime is24 hours from creation; consumed/expired10-minute invitations are not required for later selection/rematches. Browser control channels and the20-second countdown remain frontend responsibilities.

## Begin and lock

Only host can begin. The first request is:

```json
{"version":"<exact-approved-build>","action":"begin","epoch":"<fresh-UUID>","previousEpoch":"","previousMatchId":""}
```

For another round, include the current selection epoch and its terminal server match ID in the two previous fields. The previous match must be confirmed/disputed/aborted (existing terminal rules, including normal timeout handling). Active or pending blocks selection. The same current epoch with the identical body returns the current view without resetting locks, including after allocation moves `currentMatch`. Changed bodies, older/recycled epochs and mismatched previous associations return409. An epoch cannot be reused during the room lifetime. Beginning a new epoch resets both room heroes to1 and locks to false, while earlier match snapshots/reports remain immutable.

Each participant locks only their own authenticated seat:

```json
{"version":"<exact-approved-build>","action":"lock","epoch":"<current-UUID>","hero":3}
```

Hero must belong to the exact accepted roster. Unknown fields, seat overrides and opponent heroes are rejected. Identical locks are idempotent; changing a locked hero, unlocking or stale epochs fails. Exact same-hero retries remain safe after match allocation to recover a lost response; they never modify a consumed selection. Host cannot lock guest. Announce readiness only after the successful server response; countdown expiry still requires each participant's own request/token.

Both operations and GET room return the existing RoomView plus optional `selection:{epoch,previousEpoch,previousMatchId,locked:[bool,bool]}`. Internal consumed match markers and token hashes are hidden. New-version join retry with the original hero1 seat reservation remains valid after the participant selected another hero.

## Match allocation and readiness

Host uses a fresh request ID each round:

```json
{"requestId":"<fresh-round-request-id>","version":"<exact-approved-build>","selectionEpoch":"<current-UUID>"}
```

Require both server locks and the current unconsumed epoch. Atomically consume the epoch and snapshot the locked heroes into `match.players`. Different request IDs cannot allocate another match for that epoch even after completion. Exact request replay returns the same match; changed body returns409. Optional `selectionEpoch` is echoed on create/read/ready/result views; omitted fields retain legacy JSON/digests. Each participant still calls the existing `/matches/{id}/ready` with `{version}`; only both acknowledgements produce `in_progress`. Result bodies, reconciliation and trust cohorts remain unchanged. For rematch, retain room/session/P2P, finish the previous reports, begin a fresh epoch, lock again and allocate a fresh match.

## Casual network policy

For an explicitly enabled new build, the exact extra accepted seven-field tuple is:

```json
{"direction":"above","rttMs":500,"jitterMs":250,"lossPct":30,"minSamples":1,"window":12,"maxAgeMs":10000}
```

Legacy validation remains accepted for old clients. The new tuple is echoed exactly; arbitrary variants fail. These values describe browser warning/sample settings, not backend quality vetoes. Backend never waits for24 samples or simulates network/combat frames. Browser distinguishes live hints from genuine reliable-control disconnect/congestion. No physical5G acceptance or competitive anti-cheat is claimed: colluding P2P peers can still forge matching histories.

## Candidate build and release boundary

Use a reviewed roster manifest and features manifest explicitly; the helper builds current source with temporary Go overlays, cached dependencies and bounded concurrency:

```sh
python3 scripts/build-profile.py --rosters '<REVIEWED_ROSTERS_JSON>' --features '<REVIEWED_FEATURES_JSON>' --output bin/dueld-candidate
```

Add `--goos linux --goarch amd64` for a deployment artifact. This performs no database access, starts no service and does not publish. The earlier `build-production-v13.py` remains a historical pinned reproduction helper and deliberately rejects changed runtime sources; use its original source checkout for rollback reproduction. Never repin old binary digests to hide new runtime code.

Selection/match association lives in existing JSON payloads: migrations1–4, historical data and runtime database privileges stay unchanged. Before activation, freeze the exact frontend runtime, bind it in both manifests and existing SQL roster version metadata, pass isolated API/browser acceptance, then coordinate deployment to the existing API listener/origin. Do not expose another public listener or create credentials. Rollback restores the previous binary/profile; added JSON fields can remain in historical rows. An older server does not enforce the new selection gates, so the new frontend must be rolled back or disabled at the same time. No production build is authorized by a local-only QA label.
