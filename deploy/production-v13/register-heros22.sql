-- Exact reviewed frontend4608b48; additive metadata only, no old roster rewrites.
START TRANSACTION;
INSERT INTO dota_duel.duel_gameplay_rosters(roster_id,registry_version,registry_sha256,game_versions) VALUES('arena-heros22-v1','duel-heroes-127-v1','5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138',JSON_ARRAY('duel-27c78aa4cfc8facc8a23'));
SELECT ROW_COUNT();
INSERT INTO dota_duel.duel_roster_heroes(roster_id,hero_id) VALUES ('arena-heros22-v1',1),('arena-heros22-v1',3),('arena-heros22-v1',4),('arena-heros22-v1',5),('arena-heros22-v1',7),('arena-heros22-v1',8),('arena-heros22-v1',9),('arena-heros22-v1',15),('arena-heros22-v1',17),('arena-heros22-v1',18),('arena-heros22-v1',28),('arena-heros22-v1',31),('arena-heros22-v1',32),('arena-heros22-v1',36),('arena-heros22-v1',50),('arena-heros22-v1',55),('arena-heros22-v1',57),('arena-heros22-v1',58),('arena-heros22-v1',62),('arena-heros22-v1',71),('arena-heros22-v1',81),('arena-heros22-v1',82);
SELECT ROW_COUNT();
COMMIT;
