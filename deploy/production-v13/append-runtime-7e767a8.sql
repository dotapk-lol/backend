-- Approved UI-only frontend commit43bf16d1006468d8d630d14b2a7ed91022c6a77a.
-- Compare-and-swap only the existing46 version list; no hero/identity/grant/schema change.
START TRANSACTION;
UPDATE dota_duel.duel_gameplay_rosters
SET game_versions=JSON_ARRAY_APPEND(game_versions,'$','duel-7e767a8ed2b995465875')
WHERE roster_id='arena-first22-46-v1'
 AND registry_version='duel-heroes-127-v1'
 AND registry_sha256='5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138'
 AND game_versions=JSON_ARRAY('duel-e63dafb5ae2070a90f8b');
SELECT ROW_COUNT();
COMMIT;
