-- Frozen DOTA PK frontend 913f87ee; final material-epoch runtime metadata only.
START TRANSACTION;
UPDATE dota_duel.duel_gameplay_rosters SET game_versions=JSON_ARRAY('duel-27c78aa4cfc8facc8a23','duel-6b1d12f75aa4bbac4e12','duel-851e67d77307f479f1fa','duel-9431984810f197b393c5','duel-2f81eeda15fb572139ad','duel-2c2ad50b276ef596598c','duel-64714553cccdaf8cc362') WHERE roster_id='arena-heros22-v1' AND game_versions=JSON_ARRAY('duel-27c78aa4cfc8facc8a23','duel-6b1d12f75aa4bbac4e12','duel-851e67d77307f479f1fa','duel-9431984810f197b393c5','duel-2f81eeda15fb572139ad','duel-2c2ad50b276ef596598c') AND registry_version='duel-heroes-127-v1' AND registry_sha256='5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138';
SELECT ROW_COUNT();
COMMIT;
