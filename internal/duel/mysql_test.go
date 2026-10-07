package duel

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

func TestMySQLIntegration(t *testing.T) {
	dsn := os.Getenv("DUEL_TEST_DSN")
	if dsn == "" {
		t.Skip("set DUEL_TEST_DSN to an isolated Unix-socket test database")
	}
	c, e := mysql.ParseDSN(dsn)
	if e != nil || c.Addr == "/tmp/duel-mysql-test-20261002/mysql.sock" || c.Addr == "/tmp/duel-mysql-test-registry-b0-20261002/mysql.sock" || c.Net != "unix" || (!strings.Contains(c.Addr, "duel-mysql-test-") && !strings.Contains(c.Addr, ".mysql-integration")) || c.DBName != "dota_duel" {
		t.Fatal("test reset allowed only on explicit task-owned temporary Unix socket")
	}
	store, e := OpenMySQL(dsn)
	store = must(t, store, e)
	defer store.DB.Close()
	if e = store.Ping(ctx); e != nil {
		t.Fatal(e)
	}
	for _, file := range []string{"../../migrations/001_init.sql", "../../migrations/002_analytics.sql", "../../migrations/003_local_pvp_analytics.sql", "../../migrations/004_hero_registry.sql"} {
		if err := applyTestMigration(store, file); err != nil {
			t.Fatalf("migration %s: %v", file, err)
		}
	}
	factory := func(t *testing.T) Store {
		for _, table := range tables {
			if _, e := store.DB.Exec("DELETE FROM " + table); e != nil {
				t.Fatal(e)
			}
		}
		return store
	}
	runSuite(t, factory)
	runLocalSuite(t, factory)
	runReconciliationSuite(t, factory)
	runSelectionSuite(t, factory)
	runRegistrySuite(t, factory)
	runRegistrySQL(t, store, factory)
	t.Run("sql-cleanup-and-analysis", func(t *testing.T) {
		s := NewService(factory(t))
		now := time.Now()
		s.Now = func() time.Time { return now }
		a, b, r := pair(t, s)
		m := start(t, s, a, b, r, "match_request_0001")
		_, e := s.Submit(ctx, a.Token, m.ID, completed())
		if e != nil {
			t.Fatal(e)
		}
		_, e = s.Submit(ctx, b.Token, m.ID, completed())
		if e != nil {
			t.Fatal(e)
		}
		pve, e := s.CreatePVE(ctx, a.Token, PVERequest{"pve_request_00001", "duel-test", 0, 3, "normal", ""})
		pve = must(t, pve, e)
		_, e = s.Submit(ctx, a.Token, pve.ID, completed())
		if e != nil {
			t.Fatal(e)
		}
		active := start(t, s, a, b, r, "match_request_0002")
		_, e = s.Submit(ctx, a.Token, active.ID, completed())
		if e != nil {
			t.Fatal(e)
		}
		now = now.Add(RoomTTL + time.Second)
		if e = store.Cleanup(context.Background(), now.UnixMilli()); e != nil {
			t.Fatal(e)
		}
		late, e := s.GetMatch(ctx, a.Token, active.ID)
		late = must(t, late, e)
		if late.Status != "aborted" {
			t.Fatal(late.Status)
		}
		rv, e := s.GetRoom(ctx, a.Token, r.ID)
		rv = must(t, rv, e)
		if rv.Offer != nil || rv.Answer != nil {
			t.Fatal("expired SDP retained")
		}
		var count int
		if e = store.DB.QueryRow("SELECT SUM(appearances) FROM duel_hero_balance WHERE mode='pvp'").Scan(&count); e != nil || count != 2 {
			t.Fatal("PVP aggregation", count, e)
		}
		if e = store.DB.QueryRow("SELECT SUM(appearances) FROM duel_hero_balance WHERE mode='pve'").Scan(&count); e != nil || count != 1 {
			t.Fatal("PVE aggregation", count, e)
		}
		if e = store.DB.QueryRow("SELECT SUM(invalid_completed_winner+missing_confirmation+invalid_time_order) FROM duel_data_quality").Scan(&count); e != nil || count != 0 {
			t.Fatal("data integrity", count, e)
		}
		// Add two client-reported local PVP games. Legacy aggregation stays strict.
		for _, transport := range []string{"local", "broadcastchannel"} {
			input := localInput(transport)
			input.RequestID = "local_stats_" + transport
			local, err := s.CreateLocal(ctx, a.Token, input)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.Submit(ctx, a.Token, local.ID, completed()); err != nil {
				t.Fatal(err)
			}
		}
		if e = store.DB.QueryRow("SELECT SUM(appearances) FROM duel_hero_balance WHERE mode='pvp'").Scan(&count); e != nil || count != 2 {
			t.Fatal("legacy aggregate changed", count, e)
		}
		for _, transport := range []string{"local", "broadcastchannel"} {
			if e = store.DB.QueryRow("SELECT SUM(appearances) FROM duel_hero_balance_v2 WHERE mode='pvp' AND trust='client_reported' AND transport=?", transport).Scan(&count); e != nil || count != 2 {
				t.Fatal("local cohort missing", transport, count, e)
			}
		}
		if e = store.DB.QueryRow("SELECT SUM(appearances) FROM duel_hero_balance_v2 WHERE mode='pvp' AND trust='peer_agreement'").Scan(&count); e != nil || count != 2 {
			t.Fatal("trust cohorts merged", count, e)
		}
		if e = store.DB.QueryRow("SELECT SUM(invalid_trust_promotion) FROM duel_data_quality_v2").Scan(&count); e != nil || count != 0 {
			t.Fatal("promoted unverified result", count, e)
		}
		// Expired SDP removal must not prevent a rematch in an already-established P2P room.
		_ = start(t, s, a, b, r, "match_request_0003")
	})
}
