[English](README.md) | [简体中文](README.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# Pinned gameplay profile

This file set preserves explicit legacy20, historical `arena-first22-46-v1` and released `arena-heros22-v1` bindings in [gameplay-rosters.json](gameplay-rosters.json). It does not unlock46/127 in the current frontend, which remains22/88. Default source embed remains legacy20.

For the pinned reproduction helper:

| Input | Identity |
| --- | --- |
| Go source commit | `aa04e01f25ef2226b213cfc64e68c6c96fd0e18c` |
| Profile SHA256 | `405af9bf209f8d343e4d8e3dd8f2beaa88ec22b4fe06fe8b64e0a5785bad592b` |
| Expected Linux binary SHA256 | `94e0bbb7bd6bddefab84e15fd3ff7db53f47245813dd0a826ca50088d637f5ca` |
| Toolchain | Go1.26.1, CGO off, linux/amd64, cached modules |

From root, `python3 scripts/build-production-v13.py` checks actual working source against that commit and reproduces the artifact with this overlay. No database/production connection or restart occurs. A mismatch requires investigation, not changing expected digests. This is historical artifact identity, not an uptime claim.

Migration004 seeds only legacy20. `register-roster.sql` adds the reviewed historical46; `register-heros22.sql` adds22; append-runtime files preserve the reviewed exact build lists. Registration requires matching schema preconditions; INSERTs are not safe to repeat blindly. Runtime profile and SQL metadata must agree. Never use wildcard builds or treat catalog array position as identity.

`heros22_profile_test.go` is applied by `scripts/test-production-v13.py` in memory, not a game/browser test. Local use, operational review and rollback are described in [development](../../docs/DEVELOPMENT.md), [deployment](../../docs/DEPLOYMENT.md), [testing](../../docs/TESTING.md). Historical per-game/provision receipts are not retained in the working tree; original Git history remains.
