-- Reviewed frontend d81c67b; shared release profile and strict entry roster validation.
-- Preserve the already-registered6750 identity and every historical roster/member.
START TRANSACTION;
UPDATE dota_duel.duel_gameplay_rosters SET game_versions=JSON_ARRAY_APPEND(game_versions,'$','duel-1518053e640d9203bbf2') WHERE roster_id='arena-heros28-v1' AND registry_version='duel-heroes-127-v1' AND registry_sha256='5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138' AND game_versions=JSON_ARRAY('duel-6750ae2e1cfd1c3cdd42');
SELECT ROW_COUNT();
COMMIT;
