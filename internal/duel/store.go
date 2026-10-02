package duel

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Tx interface {
	Get(kind, key string, out any) error
	Put(kind, key string, value any, expires int64) error
	Insert(kind, key string, value any, expires int64) error
	Delete(kind, key string) error
}
type Store interface {
	Run(context.Context, func(Tx) error) error
	Cleanup(context.Context, int64) error
	Ping(context.Context) error
}
type SQLStore struct{ DB *sql.DB }

var tables = map[string]string{"session": "duel_sessions", "room": "duel_rooms", "code": "duel_codes", "match": "duel_matches", "request": "duel_requests", "rate": "duel_limits"}

type sqlTx struct {
	tx  *sql.Tx
	ctx context.Context
}

func table(k string) string {
	t, ok := tables[k]
	if !ok {
		panic("invalid table")
	}
	return t
}
func (s *SQLStore) Run(ctx context.Context, fn func(Tx) error) error {
	for attempt := 0; attempt < 5; attempt++ {
		tx, e := s.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if e != nil {
			return e
		}
		e = fn(&sqlTx{tx, ctx})
		if e == nil {
			e = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
		var me *mysql.MySQLError
		if errors.As(e, &me) && (me.Number == 1213 || me.Number == 1205) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 10 * time.Millisecond):
			}
			continue
		}
		return e
	}
	return errors.New("transaction retry exhausted")
}
func (s *SQLStore) Ping(ctx context.Context) error { return s.DB.PingContext(ctx) }
func (s *SQLStore) Cleanup(ctx context.Context, now int64) error {
	// Never delete match evidence. Expired SDP is removed after the room lifetime.
	for _, k := range []string{"code", "session", "request", "rate"} {
		if _, e := s.DB.ExecContext(ctx, "DELETE FROM "+table(k)+" WHERE expires_at <= ? LIMIT 500", now); e != nil {
			return e
		}
	}
	_, e := s.DB.ExecContext(ctx, `UPDATE duel_rooms SET payload=JSON_SET(payload,'$.offer',NULL,'$.answer',NULL) WHERE expires_at <= ? AND (JSON_TYPE(JSON_EXTRACT(payload,'$.offer')) <> 'NULL' OR JSON_TYPE(JSON_EXTRACT(payload,'$.answer')) <> 'NULL') LIMIT 500`, now)
	if e != nil {
		return e
	}
	// Persist timeout state even when neither client returns.
	_, e = s.DB.ExecContext(ctx, `UPDATE duel_matches SET payload=JSON_SET(payload,'$.status','aborted','$.scoreAgreement','unresolved','$.reason',IF(JSON_EXTRACT(payload,'$.startedAt')=0,'start_timeout','result_timeout'),'$.endedAt',?) WHERE status IN ('awaiting_ready','in_progress','pending') AND expires_at <= ? LIMIT 500`, now, now)
	return e
}
func (t *sqlTx) Get(k, id string, out any) error {
	var b []byte
	e := t.tx.QueryRowContext(t.ctx, "SELECT payload FROM "+table(k)+" WHERE id=? FOR UPDATE", id).Scan(&b)
	if errors.Is(e, sql.ErrNoRows) {
		return ErrMissing
	}
	if e != nil {
		return e
	}
	return json.Unmarshal(b, out)
}
func (t *sqlTx) write(k, id string, v any, expires int64, insert bool) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	q := "INSERT INTO " + table(k) + " (id,payload,expires_at) VALUES (?,?,?)"
	if !insert {
		q += " ON DUPLICATE KEY UPDATE payload=VALUES(payload),expires_at=VALUES(expires_at)"
	}
	_, e = t.tx.ExecContext(t.ctx, q, id, b, expires)
	var me *mysql.MySQLError
	if errors.As(e, &me) && me.Number == 1062 {
		return ErrExists
	}
	return e
}
func (t *sqlTx) Put(k, id string, v any, expires int64) error {
	return t.write(k, id, v, expires, false)
}
func (t *sqlTx) Insert(k, id string, v any, expires int64) error {
	return t.write(k, id, v, expires, true)
}
func (t *sqlTx) Delete(k, id string) error {
	_, e := t.tx.ExecContext(t.ctx, "DELETE FROM "+table(k)+" WHERE id=?", id)
	return e
}
func OpenMySQL(dsn string) (*SQLStore, error) {
	c, e := mysql.ParseDSN(dsn)
	if e != nil {
		return nil, errors.New("invalid MySQL DSN")
	}
	if c.DBName != "dota_duel" {
		return nil, fmt.Errorf("database must be dedicated dota_duel schema")
	}
	c.ParseTime = true
	c.Timeout = 5 * time.Second
	c.ReadTimeout = 5 * time.Second
	c.WriteTimeout = 5 * time.Second
	c.MultiStatements = false
	db, e := sql.Open("mysql", c.FormatDSN())
	if e != nil {
		return nil, errors.New("cannot open MySQL")
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(3 * time.Minute)
	return &SQLStore{db}, nil
}
