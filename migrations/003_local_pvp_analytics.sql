USE dota_duel;
-- Additive migration: keep old views and old rows unchanged for existing callers.
-- Consumers must retain trust + transport dimensions when comparing win rates.
CREATE OR REPLACE VIEW duel_hero_balance_v2 AS
SELECT game_version, mode, status,
 JSON_UNQUOTE(JSON_EXTRACT(payload,'$.trust')) AS trust,
 COALESCE(JSON_UNQUOTE(JSON_EXTRACT(payload,'$.transport')),IF(mode='pvp','webrtc','local')) AS transport,
 JSON_UNQUOTE(JSON_EXTRACT(payload,'$.aiDifficulty')) AS ai_difficulty,
 seats.seat,
 CAST(JSON_EXTRACT(payload,CONCAT('$.players[',seats.seat,'].hero')) AS UNSIGNED) AS hero,
 CAST(JSON_EXTRACT(payload,CONCAT('$.players[',1-seats.seat,'].hero')) AS UNSIGNED) AS opponent_hero,
 COUNT(*) AS appearances, SUM(winner=seats.seat) AS wins,
 ROUND(AVG(winner=seats.seat),4) AS win_rate,
 ROUND(AVG(ended_at-started_at)) AS mean_duration_ms
FROM duel_matches CROSS JOIN (SELECT 0 AS seat UNION ALL SELECT 1) seats
WHERE (mode='pvp' AND status='confirmed' AND JSON_UNQUOTE(JSON_EXTRACT(payload,'$.trust'))='peer_agreement')
 OR (mode='pvp' AND status='recorded' AND JSON_UNQUOTE(JSON_EXTRACT(payload,'$.trust'))='client_reported')
 OR (mode='pve' AND status='recorded' AND seats.seat=0 AND JSON_UNQUOTE(JSON_EXTRACT(payload,'$.trust'))='client_reported')
GROUP BY game_version,mode,status,trust,transport,ai_difficulty,seats.seat,hero,opponent_hero;
CREATE OR REPLACE VIEW duel_data_quality_v2 AS
SELECT game_version, mode, status,
 JSON_UNQUOTE(JSON_EXTRACT(payload,'$.trust')) AS trust,
 COALESCE(JSON_UNQUOTE(JSON_EXTRACT(payload,'$.transport')),IF(mode='pvp','webrtc','local')) AS transport,
 COUNT(*) AS matches,
 SUM(started_at=0) AS never_started,
 SUM(ended_at>0 AND ended_at<started_at) AS invalid_time_order,
 SUM(status IN ('confirmed','recorded') AND winner NOT IN (0,1)) AS invalid_completed_winner,
 SUM(status='confirmed' AND (JSON_TYPE(JSON_EXTRACT(payload,'$.submissions[0]'))='NULL' OR JSON_TYPE(JSON_EXTRACT(payload,'$.submissions[1]'))='NULL')) AS missing_confirmation,
 SUM(status='confirmed' AND JSON_UNQUOTE(JSON_EXTRACT(payload,'$.trust'))<>'peer_agreement') AS invalid_trust_promotion
FROM duel_matches GROUP BY game_version,mode,status,trust,transport;
INSERT IGNORE INTO duel_schema_migrations(version) VALUES (3);
