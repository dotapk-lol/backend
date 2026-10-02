package duel

import (
	"os"
	"strings"
	"testing"
)

func applyTestMigration(store *SQLStore, file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	clean := []string{}
	for _, line := range strings.Split(string(b), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			clean = append(clean, line)
		}
	}
	for _, stmt := range strings.Split(strings.Join(clean, "\n"), ";") {
		if strings.TrimSpace(stmt) != "" {
			if _, err = store.DB.Exec(stmt); err != nil {
				return err
			}
		}
	}
	return nil
}

func runRegistrySQL(t *testing.T, store *SQLStore, factory func(*testing.T) Store) {
	t.Run("accepted-subset-sql-identities", func(t *testing.T) {
		s := NewService(factory(t))
		s.registry = subsetRegistry(t)
		// Test-only acceptance metadata, never present in the shipped migration.
		if _, err := store.DB.Exec("INSERT INTO duel_gameplay_rosters VALUES('test-subset-v1',?,?,JSON_ARRAY('duel-subset-test'))", RegistryVersion, RegistrySHA256); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if _, err := store.DB.Exec("DELETE FROM duel_roster_heroes WHERE roster_id='test-subset-v1'"); err != nil {
				t.Error(err)
			}
			if _, err := store.DB.Exec("DELETE FROM duel_gameplay_rosters WHERE roster_id='test-subset-v1'"); err != nil {
				t.Error(err)
			}
		})
		if _, err := store.DB.Exec("INSERT INTO duel_roster_heroes VALUES('test-subset-v1',0),('test-subset-v1',20),('test-subset-v1',126)"); err != nil {
			t.Fatal(err)
		}
		a := creds(t, s)
		m, err := s.CreatePVE(ctx, a.Token, PVERequest{"pve_request_00001", "duel-subset-test", 20, 126, "normal", "test-subset-v1"})
		m = must(t, m, err)
		result := completed()
		result.Version = m.Version
		if _, err = s.Submit(ctx, a.Token, m.ID, result); err != nil {
			t.Fatal(err)
		}
		var hero, opponent, valve, opponentValve, count, unmapped int
		if err = store.DB.QueryRow("SELECT hero,opponent_hero,hero_valve_id,opponent_valve_id,appearances,unmapped_hero+unmapped_opponent FROM duel_hero_balance_v3 WHERE roster_id='test-subset-v1'").Scan(&hero, &opponent, &valve, &opponentValve, &count, &unmapped); err != nil || hero != 20 || opponent != 126 || valve != 3 || opponentValve != 155 || count != 1 || unmapped != 0 {
			t.Fatal("expanded identities or roster joins incorrect", hero, opponent, valve, opponentValve, count, unmapped, err)
		}
	})
	t.Run("registry-seed-equals-frozen-manifest", func(t *testing.T) {
		r := mustRegistry()
		var count int
		if err := store.DB.QueryRow("SELECT COUNT(*) FROM duel_heroes").Scan(&count); err != nil || count != len(r.heroes) {
			t.Fatal(count, err)
		}
		for _, h := range r.heroes {
			var internal, key string
			var valve int
			if err := store.DB.QueryRow("SELECT internal_id,valve_id,valve_key FROM duel_heroes WHERE hero_id=?", h.HeroID).Scan(&internal, &valve, &key); err != nil || internal != h.InternalID || valve != h.ValveID || key != h.ValveKey {
				t.Fatal("SQL mapping differs", h, err)
			}
		}
		if err := store.DB.QueryRow("SELECT COUNT(*) FROM duel_roster_heroes").Scan(&count); err != nil || count != 20 {
			t.Fatal("catalog wrongly activated", count, err)
		}
		var sha string
		if err := store.DB.QueryRow("SELECT registry_sha256 FROM duel_gameplay_rosters WHERE roster_id=?", LegacyRosterID).Scan(&sha); err != nil || sha != RegistrySHA256 {
			t.Fatal("seed provenance mismatch", sha, err)
		}
	})
	t.Run("additive-migration-historical-rows-and-cohorts", func(t *testing.T) {
		s := NewService(factory(t))
		a, b, r := pair(t, s)
		m := start(t, s, a, b, r, "match_request_0001")
		if _, err := s.Submit(ctx, a.Token, m.ID, completed()); err != nil {
			t.Fatal(err)
		}
		m, err := s.Submit(ctx, b.Token, m.ID, completed())
		m = must(t, m, err)
		// A complete pre-registry payload with immutable reports; migration must not rewrite it.
		if _, err = store.DB.Exec("UPDATE duel_matches SET payload=JSON_REMOVE(payload,'$.rosterId','$.registryVersion') WHERE id=?", m.ID); err != nil {
			t.Fatal(err)
		}
		var before, after string
		if err = store.DB.QueryRow("SELECT SHA2(CAST(payload AS CHAR),256) FROM duel_matches WHERE id=?", m.ID).Scan(&before); err != nil {
			t.Fatal(err)
		}
		if err = applyTestMigration(store, "../../migrations/004_hero_registry.sql"); err != nil {
			t.Fatal(err)
		}
		if err = store.DB.QueryRow("SELECT SHA2(CAST(payload AS CHAR),256) FROM duel_matches WHERE id=?", m.ID).Scan(&after); err != nil || before != after {
			t.Fatal("migration rewrote historical payload", err)
		}
		for _, view := range []string{"duel_hero_balance", "duel_hero_balance_v2", "duel_hero_balance_v3"} {
			var count int
			if err = store.DB.QueryRow("SELECT SUM(appearances) FROM " + view).Scan(&count); err != nil || count != 2 {
				t.Fatal("fanout or dropped record", view, count, err)
			}
		}
		var roster, registry, internal string
		var valve int
		if err = store.DB.QueryRow("SELECT roster_id,registry_version,hero_internal_id,hero_valve_id FROM duel_hero_balance_v3 WHERE hero=0").Scan(&roster, &registry, &internal, &valve); err != nil || roster != LegacyRosterID || registry != LegacyRosterID || internal != "juggernaut" || valve != 8 {
			t.Fatal("legacy mapping reinterpreted", roster, registry, internal, valve, err)
		}
		for _, transport := range []string{"local", "broadcastchannel", "pve"} {
			var n Match
			if transport == "pve" {
				n, err = s.CreatePVE(ctx, a.Token, PVERequest{"pve_request_00001", "duel-test", 0, 3, "normal", LegacyRosterID})
			} else {
				in := localInput(transport)
				in.RequestID = "registry_stats_" + transport
				in.RosterID = LegacyRosterID
				n, err = s.CreateLocal(ctx, a.Token, in)
			}
			n = must(t, n, err)
			if _, err = s.Submit(ctx, a.Token, n.ID, completed()); err != nil {
				t.Fatal(err)
			}
		}
		pending := start(t, s, a, b, r, "match_request_0002")
		if _, err = s.Submit(ctx, a.Token, pending.ID, completed()); err != nil {
			t.Fatal(err)
		}
		for _, cohort := range []struct {
			mode, transport, trust string
			count                  int
		}{{"pvp", "webrtc", "peer_agreement", 2}, {"pvp", "local", "client_reported", 2}, {"pvp", "broadcastchannel", "client_reported", 2}, {"pve", "local", "client_reported", 1}} {
			var count int
			if err = store.DB.QueryRow("SELECT SUM(appearances) FROM duel_hero_balance_v3 WHERE mode=? AND transport=? AND trust=?", cohort.mode, cohort.transport, cohort.trust).Scan(&count); err != nil || count != cohort.count {
				t.Fatal("cohort regression", cohort, count, err)
			}
		}
		var issues int
		if err = store.DB.QueryRow("SELECT SUM(unknown_roster+unplayable_participant+invalid_trust_promotion) FROM duel_data_quality_v3").Scan(&issues); err != nil || issues != 0 {
			t.Fatal("unexpected quality issues", issues, err)
		}
		// Corrupt/imported records remain visible with flags, never dropped by inner joins.
		corrupt := m
		corrupt.ID = randomID()
		corrupt.Players[0].Hero = 20
		corrupt.Status = "aborted"
		if err = store.Run(ctx, func(tx Tx) error { return tx.Insert("match", corrupt.ID, corrupt, corrupt.Deadline) }); err != nil {
			t.Fatal(err)
		}
		if err = store.DB.QueryRow("SELECT SUM(unplayable_participant) FROM duel_data_quality_v3").Scan(&issues); err != nil || issues != 1 {
			t.Fatal("catalog hero accepted in wrong roster", issues, err)
		}
		var count int
		if err = store.DB.QueryRow("SELECT COUNT(*) FROM duel_match_heroes_v3 WHERE match_id=?", corrupt.ID).Scan(&count); err != nil || count != 2 {
			t.Fatal("unmapped record dropped", count, err)
		}
		if err = store.DB.QueryRow("SELECT SUM(appearances) FROM duel_hero_balance_v3").Scan(&count); err != nil || count != 7 {
			t.Fatal("pending/aborted counted or fanout", count, err)
		}
	})
}
