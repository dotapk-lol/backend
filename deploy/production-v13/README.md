# Frozen production v1.3 profile

This profile is the deployed legacy20 + `arena-first22-46-v1` acceptance, with sole46 build `duel-e63dafb5ae2070a90f8b`. Source default remains legacy20; no catalog127, local24 or later candidate activation is implied.

`build-metadata.json` identifies the pinned Go source commit, profile SHA and exact running Linux binary SHA. From the repository root, `python3 scripts/build-production-v13.py` reproduces that immutable source snapshot plus this profile using Go1.26.1 and the existing module cache, without database or production access. It does not deploy/restart anything.

`migrations/004_hero_registry.sql` only seeds legacy20. `register-roster.sql` is the separately reviewed46 registration applied after004 on the dedicated production schema. Production already has both; do not reapply this plain INSERT script or run fixture registration/reset scripts there. No provisioning, GRANT, credential, firewall or service change is needed.

Full source registry SHA and the minimal public/SQL identity projection hashes are different semantic objects. Exact profile order/IDs and build binding are preserved; do not infer identity from array position. Rollback and point-in-time production acceptance are documented in `docs/DEPLOYMENT.md` and `docs/production-v13-*.json`. Sensitive full backups remain outside the repository under protected server directories.
