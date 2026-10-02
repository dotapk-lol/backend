-- Additive catalog and accepted gameplay membership. No match or report is rewritten.
-- Runtime access remains DML-only on dota_duel; use the existing migration operator.
-- Catalog membership alone NEVER makes a hero playable. Only roster membership does.
USE dota_duel;
CREATE TABLE IF NOT EXISTS duel_heroes (
 hero_id INT NOT NULL PRIMARY KEY,
 internal_id VARCHAR(100) CHARACTER SET ascii COLLATE ascii_bin NOT NULL UNIQUE,
 valve_id INT NOT NULL UNIQUE,
 valve_key VARCHAR(100) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 CHECK (hero_id >= 0), CHECK (valve_id > 0)
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS duel_gameplay_rosters (
 roster_id VARCHAR(100) CHARACTER SET ascii COLLATE ascii_bin NOT NULL PRIMARY KEY,
 registry_version VARCHAR(100) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 registry_sha256 CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 game_versions JSON NOT NULL
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS duel_roster_heroes (
 roster_id VARCHAR(100) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 hero_id INT NOT NULL,
 PRIMARY KEY(roster_id,hero_id),
 FOREIGN KEY(roster_id) REFERENCES duel_gameplay_rosters(roster_id),
 FOREIGN KEY(hero_id) REFERENCES duel_heroes(hero_id)
) ENGINE=InnoDB;
-- An existing ID with a different mapping must fail, never silently remap history.
INSERT INTO duel_heroes(hero_id,internal_id,valve_id,valve_key) VALUES
(0,'juggernaut',8,'juggernaut'),
(1,'crystal_maiden',5,'crystal_maiden'),
(2,'pudge',14,'pudge'),
(3,'axe',2,'axe'),
(4,'sniper',35,'sniper'),
(5,'anti_mage',1,'antimage'),
(6,'phantom_assassin',44,'phantom_assassin'),
(7,'drow_ranger',6,'drow_ranger'),
(8,'lina',25,'lina'),
(9,'lion',26,'lion'),
(10,'earthshaker',7,'earthshaker'),
(11,'mirana',9,'mirana'),
(12,'sven',18,'sven'),
(13,'zeus',22,'zuus'),
(14,'windranger',21,'windrunner'),
(15,'shadow_fiend',11,'nevermore'),
(16,'storm_spirit',17,'storm_spirit'),
(17,'queen_of_pain',39,'queenofpain'),
(18,'witch_doctor',30,'witch_doctor'),
(19,'tidehunter',29,'tidehunter'),
(20,'valve_3',3,'bane'),
(21,'valve_4',4,'bloodseeker'),
(22,'valve_10',10,'morphling'),
(23,'valve_12',12,'phantom_lancer'),
(24,'valve_13',13,'puck'),
(25,'valve_15',15,'razor'),
(26,'valve_16',16,'sand_king'),
(27,'valve_19',19,'tiny'),
(28,'valve_20',20,'vengefulspirit'),
(29,'valve_23',23,'kunkka'),
(30,'valve_27',27,'shadow_shaman'),
(31,'valve_28',28,'slardar'),
(32,'valve_31',31,'lich'),
(33,'valve_32',32,'riki'),
(34,'valve_33',33,'enigma'),
(35,'valve_34',34,'tinker'),
(36,'valve_36',36,'necrolyte'),
(37,'valve_37',37,'warlock'),
(38,'valve_38',38,'beastmaster'),
(39,'valve_40',40,'venomancer'),
(40,'valve_41',41,'faceless_void'),
(41,'valve_42',42,'skeleton_king'),
(42,'valve_43',43,'death_prophet'),
(43,'valve_45',45,'pugna'),
(44,'valve_46',46,'templar_assassin'),
(45,'valve_47',47,'viper'),
(46,'valve_48',48,'luna'),
(47,'valve_49',49,'dragon_knight'),
(48,'valve_50',50,'dazzle'),
(49,'valve_51',51,'rattletrap'),
(50,'valve_52',52,'leshrac'),
(51,'valve_53',53,'furion'),
(52,'valve_54',54,'life_stealer'),
(53,'valve_55',55,'dark_seer'),
(54,'valve_56',56,'clinkz'),
(55,'valve_57',57,'omniknight'),
(56,'valve_58',58,'enchantress'),
(57,'valve_59',59,'huskar'),
(58,'valve_60',60,'night_stalker'),
(59,'valve_61',61,'broodmother'),
(60,'valve_62',62,'bounty_hunter'),
(61,'valve_63',63,'weaver'),
(62,'valve_64',64,'jakiro'),
(63,'valve_65',65,'batrider'),
(64,'valve_66',66,'chen'),
(65,'valve_67',67,'spectre'),
(66,'valve_68',68,'ancient_apparition'),
(67,'valve_69',69,'doom_bringer'),
(68,'valve_70',70,'ursa'),
(69,'valve_71',71,'spirit_breaker'),
(70,'valve_72',72,'gyrocopter'),
(71,'valve_73',73,'alchemist'),
(72,'valve_74',74,'invoker'),
(73,'valve_75',75,'silencer'),
(74,'valve_76',76,'obsidian_destroyer'),
(75,'valve_77',77,'lycan'),
(76,'valve_78',78,'brewmaster'),
(77,'valve_79',79,'shadow_demon'),
(78,'valve_80',80,'lone_druid'),
(79,'valve_81',81,'chaos_knight'),
(80,'valve_82',82,'meepo'),
(81,'valve_83',83,'treant'),
(82,'valve_84',84,'ogre_magi'),
(83,'valve_85',85,'undying'),
(84,'valve_86',86,'rubick'),
(85,'valve_87',87,'disruptor'),
(86,'valve_88',88,'nyx_assassin'),
(87,'valve_89',89,'naga_siren'),
(88,'valve_90',90,'keeper_of_the_light'),
(89,'valve_91',91,'wisp'),
(90,'valve_92',92,'visage'),
(91,'valve_93',93,'slark'),
(92,'valve_94',94,'medusa'),
(93,'valve_95',95,'troll_warlord'),
(94,'valve_96',96,'centaur'),
(95,'valve_97',97,'magnataur'),
(96,'valve_98',98,'shredder'),
(97,'valve_99',99,'bristleback'),
(98,'valve_100',100,'tusk'),
(99,'valve_101',101,'skywrath_mage'),
(100,'valve_102',102,'abaddon'),
(101,'valve_103',103,'elder_titan'),
(102,'valve_104',104,'legion_commander'),
(103,'valve_105',105,'techies'),
(104,'valve_106',106,'ember_spirit'),
(105,'valve_107',107,'earth_spirit'),
(106,'valve_108',108,'abyssal_underlord'),
(107,'valve_109',109,'terrorblade'),
(108,'valve_110',110,'phoenix'),
(109,'valve_111',111,'oracle'),
(110,'valve_112',112,'winter_wyvern'),
(111,'valve_113',113,'arc_warden'),
(112,'valve_114',114,'monkey_king'),
(113,'valve_119',119,'dark_willow'),
(114,'valve_120',120,'pangolier'),
(115,'valve_121',121,'grimstroke'),
(116,'valve_123',123,'hoodwink'),
(117,'valve_126',126,'void_spirit'),
(118,'valve_128',128,'snapfire'),
(119,'valve_129',129,'mars'),
(120,'valve_131',131,'ringmaster'),
(121,'valve_135',135,'dawnbreaker'),
(122,'valve_136',136,'marci'),
(123,'valve_137',137,'primal_beast'),
(124,'valve_138',138,'muerta'),
(125,'valve_145',145,'kez'),
(126,'valve_155',155,'largo')
ON DUPLICATE KEY UPDATE hero_id=IF(hero_id=VALUES(hero_id) AND internal_id=VALUES(internal_id) AND valve_id=VALUES(valve_id) AND valve_key=VALUES(valve_key),hero_id,NULL);
INSERT INTO duel_gameplay_rosters(roster_id,registry_version,registry_sha256,game_versions) VALUES
('legacy-20-v1','duel-heroes-127-v1','5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138','[]')
ON DUPLICATE KEY UPDATE roster_id=IF(registry_version=VALUES(registry_version) AND registry_sha256=VALUES(registry_sha256) AND game_versions=VALUES(game_versions),roster_id,NULL);
INSERT IGNORE INTO duel_roster_heroes(roster_id,hero_id) VALUES
('legacy-20-v1',0),
('legacy-20-v1',1),
('legacy-20-v1',2),
('legacy-20-v1',3),
('legacy-20-v1',4),
('legacy-20-v1',5),
('legacy-20-v1',6),
('legacy-20-v1',7),
('legacy-20-v1',8),
('legacy-20-v1',9),
('legacy-20-v1',10),
('legacy-20-v1',11),
('legacy-20-v1',12),
('legacy-20-v1',13),
('legacy-20-v1',14),
('legacy-20-v1',15),
('legacy-20-v1',16),
('legacy-20-v1',17),
('legacy-20-v1',18),
('legacy-20-v1',19);
-- Historical missing metadata is interpreted only in projections, without UPDATE.
CREATE OR REPLACE VIEW duel_match_metadata_v3 AS
SELECT m.*,
 COALESCE(NULLIF(NULLIF(JSON_UNQUOTE(JSON_EXTRACT(payload,'$.rosterId')),'null'),''),'legacy-20-v1') AS roster_id,
 COALESCE(NULLIF(NULLIF(JSON_UNQUOTE(JSON_EXTRACT(payload,'$.registryVersion')),'null'),''),'legacy-20-v1') AS registry_version,
 JSON_UNQUOTE(JSON_EXTRACT(payload,'$.trust')) AS trust,
 COALESCE(JSON_UNQUOTE(JSON_EXTRACT(payload,'$.transport')),IF(mode='pvp','webrtc','local')) AS transport,
 JSON_UNQUOTE(JSON_EXTRACT(payload,'$.aiDifficulty')) AS ai_difficulty
FROM duel_matches m;
CREATE OR REPLACE VIEW duel_match_heroes_v3 AS
SELECT m.id AS match_id,m.game_version,m.roster_id,m.registry_version,m.mode,m.status,m.trust,m.transport,m.ai_difficulty,m.started_at,m.ended_at,m.winner,seats.seat,
 CAST(JSON_EXTRACT(m.payload,CONCAT('$.players[',seats.seat,'].hero')) AS SIGNED) AS hero,
 CAST(JSON_EXTRACT(m.payload,CONCAT('$.players[',1-seats.seat,'].hero')) AS SIGNED) AS opponent_hero,
 h.internal_id AS hero_internal_id,h.valve_id AS hero_valve_id,
 o.internal_id AS opponent_internal_id,o.valve_id AS opponent_valve_id,
 (rh.hero_id IS NULL) AS unmapped_hero,(ro.hero_id IS NULL) AS unmapped_opponent
FROM duel_match_metadata_v3 m CROSS JOIN (SELECT 0 AS seat UNION ALL SELECT 1) seats
LEFT JOIN duel_roster_heroes rh ON rh.roster_id=m.roster_id AND rh.hero_id=JSON_EXTRACT(m.payload,CONCAT('$.players[',seats.seat,'].hero'))
LEFT JOIN duel_roster_heroes ro ON ro.roster_id=m.roster_id AND ro.hero_id=JSON_EXTRACT(m.payload,CONCAT('$.players[',1-seats.seat,'].hero'))
LEFT JOIN duel_heroes h ON h.hero_id=rh.hero_id
LEFT JOIN duel_heroes o ON o.hero_id=ro.hero_id;
CREATE OR REPLACE VIEW duel_hero_balance_v3 AS
SELECT game_version,roster_id,registry_version,mode,status,trust,transport,ai_difficulty,seat,hero,opponent_hero,
 hero_internal_id,hero_valve_id,opponent_internal_id,opponent_valve_id,unmapped_hero,unmapped_opponent,
 COUNT(*) AS appearances,SUM(winner=seat) AS wins,ROUND(AVG(winner=seat),4) AS win_rate,ROUND(AVG(ended_at-started_at)) AS mean_duration_ms
FROM duel_match_heroes_v3
WHERE (mode='pvp' AND status='confirmed' AND trust='peer_agreement')
 OR (mode='pvp' AND status='recorded' AND trust='client_reported')
 OR (mode='pve' AND status='recorded' AND seat=0 AND trust='client_reported')
GROUP BY game_version,roster_id,registry_version,mode,status,trust,transport,ai_difficulty,seat,hero,opponent_hero,hero_internal_id,hero_valve_id,opponent_internal_id,opponent_valve_id,unmapped_hero,unmapped_opponent;
CREATE OR REPLACE VIEW duel_data_quality_v3 AS
SELECT m.game_version,m.roster_id,m.registry_version,m.mode,m.status,m.trust,m.transport,
 COUNT(*) AS matches,SUM(m.started_at=0) AS never_started,SUM(m.ended_at>0 AND m.ended_at<m.started_at) AS invalid_time_order,
 SUM(m.status IN ('confirmed','recorded') AND m.winner NOT IN (0,1)) AS invalid_completed_winner,
 SUM(m.status='confirmed' AND (JSON_TYPE(JSON_EXTRACT(m.payload,'$.submissions[0]'))='NULL' OR JSON_TYPE(JSON_EXTRACT(m.payload,'$.submissions[1]'))='NULL')) AS missing_confirmation,
 SUM(m.status='confirmed' AND m.trust<>'peer_agreement') AS invalid_trust_promotion,
 SUM(r.roster_id IS NULL) AS unknown_roster,SUM(h.hero_id IS NULL OR o.hero_id IS NULL) AS unplayable_participant
FROM duel_match_metadata_v3 m
LEFT JOIN duel_gameplay_rosters r ON r.roster_id=m.roster_id
LEFT JOIN duel_roster_heroes h ON h.roster_id=m.roster_id AND h.hero_id=JSON_EXTRACT(m.payload,'$.players[0].hero')
LEFT JOIN duel_roster_heroes o ON o.roster_id=m.roster_id AND o.hero_id=JSON_EXTRACT(m.payload,'$.players[1].hero')
GROUP BY m.game_version,m.roster_id,m.registry_version,m.mode,m.status,m.trust,m.transport;
INSERT IGNORE INTO duel_schema_migrations(version) VALUES(4);
