[English](README.md) | [简体中文](README.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# Pinned gameplay profile

This file set preserves explicit legacy20, historical `arena-first22-46-v1` and released `arena-heros22-v1` bindings in [gameplay-rosters.json](gameplay-rosters.json). It does not unlock46/127 in the current frontend, which remains22/88. Default source embed remains legacy20.

For the pinned reproduction helper:

| Input | Identity |
| --- | --- |
| Go source commit | `aa04e01f25ef2226b213cfc64e68c6c96fd0e18c` |
| Profile SHA256 | `7fbdab24680a8b1b1101d5c596274e39a0bf757880fcf481e82da9ff5247023f` |
| Expected Linux binary SHA256 | `9512f02b72d20c97f14f8855559f42d902c6a13fb0fe8b7cd03baa859dcc1b8c` |
| Toolchain | Go1.26.1, CGO off, linux/amd64, cached modules |

From root, `python3 scripts/build-production-v13.py` checks actual working source against that commit and reproduces the artifact with this overlay. No database/production connection or restart occurs. A mismatch requires investigation, not changing expected digests. This is historical artifact identity, not an uptime claim.

Migration004 seeds only legacy20. `register-roster.sql` adds the reviewed historical46; `register-heros22.sql` adds22; append-runtime files preserve the reviewed exact build lists. Registration requires matching schema preconditions; INSERTs are not safe to repeat blindly. Runtime profile and SQL metadata must agree. Never use wildcard builds or treat catalog array position as identity.

`heros22_profile_test.go` is applied by `scripts/test-production-v13.py` in memory, not a game/browser test. Local use, operational review and rollback are described in [development](../../docs/DEVELOPMENT.md), [deployment](../../docs/DEPLOYMENT.md), [testing](../../docs/TESTING.md). Historical per-game/provision receipts are not retained in the working tree; original Git history remains.

The reviewed standard frontend build at commit `fcec6cf62c5ba53e397d3fef45e17ac9e859cb08` uses `duel-9431984810f197b393c5`. Its exact binding is appended by `append-privacy-release-runtime.sql`; all earlier22,46 and legacy bindings remain unchanged. Hero membership, rules and seven-field quality policy do not change. Deployment receipts, host configuration, backups and per-game verification remain private and outside source control. Source updates and these artifact digests do not claim frontend publication or browser acceptance.

The newer room-selection source changes runtime bytes. This pinned helper deliberately rejects it; use the original source checkout for historical reproduction and [build-profile.py](../../scripts/build-profile.py) with explicitly reviewed roster/features manifests for new candidates. See [selection release boundaries](../../docs/SELECTION.md). This existing profile retains every previously approved build.

The reviewed DOTA PK frontend73341e2 runtime `duel-2f81eeda15fb572139ad` is explicitly appended to the released22 profile and is the only version in its `protocol-features.json` roomSelectionVersions. Keep all previous bindings. `append-selection-runtime.sql` updates existing SQL metadata by exact compare-and-swap, without migrations or new privileges. Build current deployment source with `scripts/build-profile.py` using both production manifests; the earlier pinned digests remain historical. Source registration is separate from actual frontend publication.
