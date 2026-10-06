-- Exact reviewed i18n frontend e9dd04a; compare-and-swap existing22 version list.
START TRANSACTION;
UPDATE dota_duel.duel_gameplay_rosters SET game_versions=JSON_ARRAY_APPEND(game_versions,'$','duel-851e67d77307f479f1fa') WHERE roster_id='arena-heros22-v1' AND registry_version='duel-heroes-127-v1' AND registry_sha256='5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138' AND game_versions=JSON_ARRAY('duel-27c78aa4cfc8facc8a23','duel-6b1d12f75aa4bbac4e12');
SELECT ROW_COUNT();
COMMIT;
