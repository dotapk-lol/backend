-- Exact reviewed frontend 57a69c1; independent28 metadata only.
-- Additive metadata only; never expand the historical22 roster. Not safe to replay blindly.
START TRANSACTION;
INSERT INTO dota_duel.duel_gameplay_rosters(roster_id,registry_version,registry_sha256,game_versions) VALUES('arena-heros28-v1','duel-heroes-127-v1','5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138',JSON_ARRAY('duel-6750ae2e1cfd1c3cdd42'));
SELECT ROW_COUNT();
INSERT INTO dota_duel.duel_roster_heroes(roster_id,hero_id) VALUES ('arena-heros28-v1',0),('arena-heros28-v1',1),('arena-heros28-v1',2),('arena-heros28-v1',3),('arena-heros28-v1',4),('arena-heros28-v1',5),('arena-heros28-v1',6),('arena-heros28-v1',7),('arena-heros28-v1',8),('arena-heros28-v1',9),('arena-heros28-v1',10),('arena-heros28-v1',14),('arena-heros28-v1',15),('arena-heros28-v1',17),('arena-heros28-v1',18),('arena-heros28-v1',19),('arena-heros28-v1',28),('arena-heros28-v1',31),('arena-heros28-v1',32),('arena-heros28-v1',36),('arena-heros28-v1',50),('arena-heros28-v1',55),('arena-heros28-v1',57),('arena-heros28-v1',58),('arena-heros28-v1',62),('arena-heros28-v1',71),('arena-heros28-v1',81),('arena-heros28-v1',82);
SELECT ROW_COUNT();
COMMIT;
