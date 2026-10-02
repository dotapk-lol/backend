-- Apply only to the dedicated dota_duel schema. Never select AgentSquared's schema.
-- Run separately with a migration account; the runtime account needs only DML.
USE dota_duel;
CREATE TABLE IF NOT EXISTS duel_sessions (
 id VARCHAR(160) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
 payload JSON NOT NULL, expires_at BIGINT NOT NULL, INDEX expires_idx(expires_at)
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS duel_rooms LIKE duel_sessions;
CREATE TABLE IF NOT EXISTS duel_codes LIKE duel_sessions;
CREATE TABLE IF NOT EXISTS duel_requests LIKE duel_sessions;
CREATE TABLE IF NOT EXISTS duel_limits LIKE duel_sessions;
CREATE TABLE IF NOT EXISTS duel_matches (
 id VARCHAR(160) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
 payload JSON NOT NULL, expires_at BIGINT NOT NULL,
 status VARCHAR(24) GENERATED ALWAYS AS (JSON_UNQUOTE(JSON_EXTRACT(payload,'$.status'))) STORED,
 mode VARCHAR(8) GENERATED ALWAYS AS (JSON_UNQUOTE(JSON_EXTRACT(payload,'$.mode'))) STORED,
 game_version VARCHAR(100) GENERATED ALWAYS AS (JSON_UNQUOTE(JSON_EXTRACT(payload,'$.version'))) STORED,
 started_at BIGINT GENERATED ALWAYS AS (JSON_EXTRACT(payload,'$.startedAt')) STORED,
 ended_at BIGINT GENERATED ALWAYS AS (JSON_EXTRACT(payload,'$.endedAt')) STORED,
 winner INT GENERATED ALWAYS AS (JSON_EXTRACT(payload,'$.winner')) STORED,
 INDEX expiry_idx(status,expires_at), INDEX analysis_idx(game_version,mode,status), INDEX started_idx(started_at)
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS duel_schema_migrations (version INT PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
INSERT IGNORE INTO duel_schema_migrations(version) VALUES (1);
