package duel

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestFrozenRegistry(t *testing.T) {
	r := mustRegistry()
	ids := []int{8, 5, 14, 2, 35, 1, 44, 6, 25, 26, 7, 9, 18, 22, 21, 11, 17, 39, 30, 29}
	names := strings.Split("juggernaut crystal_maiden pudge axe sniper anti_mage phantom_assassin drow_ranger lina lion earthshaker mirana sven zeus windranger shadow_fiend storm_spirit queen_of_pain witch_doctor tidehunter", " ")
	if len(r.heroes) != 127 {
		t.Fatal("catalog count changed without reviewed fixture")
	}
	for i, valve := range ids {
		h := r.byID[i]
		if h.ValveID != valve || h.InternalID != names[i] || h.LegacyIndex == nil || *h.LegacyIndex != i {
			t.Fatalf("legacy identity %d changed: %+v", i, h)
		}
	}
	for i := 20; i < 127; i++ {
		h, ok := r.byID[i]
		if !ok || h.LegacyIndex != nil || h.InternalID != fmt.Sprintf("valve_%d", h.ValveID) {
			t.Fatalf("invalid appended identity %d", i)
		}
	}
	if len(r.rosters) != 1 || r.rosters[0].ID != LegacyRosterID {
		t.Fatal("unaccepted gameplay roster activated")
	}
	if _, err := r.resolve(RegistryVersion, "duel-test", 20); err == nil {
		t.Fatal("catalog enabled as gameplay roster")
	}
	if _, err := loadRegistry([]byte(strings.Replace(string(registryJSON), `"juggernaut"`, `"changed"`, 1)), gameplayJSON); err == nil {
		t.Fatal("corrupt catalog accepted")
	}
}

// This subset is only a protocol fixture. It is not shipped or claimed playable.
func subsetRegistry(t *testing.T) *heroRegistry {
	t.Helper()
	var rosters []GameplayRoster
	if err := json.Unmarshal(gameplayJSON, &rosters); err != nil {
		t.Fatal(err)
	}
	rosters = append(rosters, GameplayRoster{"test-subset-v1", RegistryVersion, []int{0, 20, 126}, []string{"duel-subset-test"}})
	b, _ := json.Marshal(rosters)
	r, err := loadRegistry(registryJSON, b)
	return must(t, r, err)
}

func TestRosterConfigurationFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name   string
		roster GameplayRoster
	}{
		{"unknown hero", GameplayRoster{"test-subset-v1", RegistryVersion, []int{127}, []string{"duel-subset-test"}}},
		{"duplicate hero", GameplayRoster{"test-subset-v1", RegistryVersion, []int{20, 20}, []string{"duel-subset-test"}}},
		{"no build acceptance", GameplayRoster{"test-subset-v1", RegistryVersion, []int{20}, nil}},
		{"catalog as roster", GameplayRoster{RegistryVersion, RegistryVersion, []int{20}, []string{"duel-subset-test"}}},
		{"wrong registry", GameplayRoster{"test-subset-v1", "other", []int{20}, []string{"duel-subset-test"}}},
		{"legacy changed", GameplayRoster{LegacyRosterID, RegistryVersion, []int{0, 20}, nil}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var rosters []GameplayRoster
			_ = json.Unmarshal(gameplayJSON, &rosters)
			if tc.roster.ID == LegacyRosterID {
				rosters = nil
			}
			rosters = append(rosters, tc.roster)
			b, _ := json.Marshal(rosters)
			if _, err := loadRegistry(registryJSON, b); err == nil {
				t.Fatal("unsafe roster accepted")
			}
		})
	}
}

func TestLegacyRequestBytes(t *testing.T) {
	// These wire snapshots predate rosterId. Their hashes are stored for 24 hours.
	for _, tc := range []struct {
		in  any
		old string
	}{
		{roomInput(), `{"requestId":"request_1234567890","version":"duel-test","hero":0,"offer":{"type":"offer","sdp":"v=0\r\n"},"policy":{"direction":"above","rttMs":200,"jitterMs":30,"lossPct":5,"minSamples":24,"window":30,"maxAgeMs":3000}}`},
		{PVERequest{"pve_request_00001", "duel-test", 0, 3, "normal", ""}, `{"requestId":"pve_request_00001","version":"duel-test","hero":0,"opponentHero":3,"aiDifficulty":"normal"}`},
		{localInput("local"), `{"requestId":"local_request_0001","version":"duel-test","hero":1,"opponentHero":4,"transport":"local"}`},
		{CreateMatch{"match_request_0001", "duel-test", ""}, `{"requestId":"match_request_0001","version":"duel-test"}`},
		{completed(), `{"version":"duel-test","outcome":"completed","rounds":[{"number":1,"winner":0,"remainingMs":1200},{"number":2,"winner":-1,"remainingMs":0},{"number":3,"winner":0,"remainingMs":2000}],"score":[2,0],"winner":0,"reason":""}`},
	} {
		b, _ := json.Marshal(tc.in)
		if string(b) != tc.old || digest(tc.in) != hash(tc.old) {
			t.Fatalf("legacy digest changed: %s", b)
		}
	}
}

func runRegistrySuite(t *testing.T, factory func(*testing.T) Store) {
	t.Run("legacy-persisted-room-replay-and-read", func(t *testing.T) {
		s := NewService(factory(t))
		a := creds(t, s)
		r, err := s.CreateRoom(ctx, a.Token, roomInput())
		r = must(t, r, err)
		// Simulate a persisted v1.2 row and idempotency record, with no metadata.
		err = s.Store.Run(ctx, func(tx Tx) error {
			var old map[string]any
			if err := tx.Get("room", r.ID, &old); err != nil {
				return err
			}
			delete(old, "rosterId")
			delete(old, "registryVersion")
			if err := tx.Put("room", r.ID, old, r.Expires); err != nil {
				return err
			}
			return tx.Put("request", reqKey("room", a.Token, roomInput().RequestID), requestRef{r.ID, hash(`{"requestId":"request_1234567890","version":"duel-test","hero":0,"offer":{"type":"offer","sdp":"v=0\r\n"},"policy":{"direction":"above","rttMs":200,"jitterMs":30,"lossPct":5,"minSamples":24,"window":30,"maxAgeMs":3000}}`)}, a.Expires)
		})
		if err != nil {
			t.Fatal(err)
		}
		same, err := s.CreateRoom(ctx, a.Token, roomInput())
		same = must(t, same, err)
		if same.ID != r.ID || same.RosterID != LegacyRosterID || same.RegistryVersion != LegacyRosterID {
			t.Fatalf("legacy replay: %+v", same)
		}
		b := creds(t, s)
		if _, err = s.JoinRoom(ctx, b.Token, JoinRoom{r.Code, "duel-test", 3, LegacyRosterID}); err != nil {
			t.Fatal(err)
		}
		if err = s.Answer(ctx, b.Token, r.ID, "duel-test", Description{"answer", "v=0\r\n"}); err != nil {
			t.Fatal(err)
		}
		m := start(t, s, a, b, r, "match_request_0001")
		if m.RosterID != LegacyRosterID || m.RegistryVersion != LegacyRosterID {
			t.Fatal("old room snapshot changed")
		}
		if _, err = s.Submit(ctx, a.Token, m.ID, completed()); err != nil {
			t.Fatal(err)
		}
		m, err = s.Submit(ctx, b.Token, m.ID, completed())
		m = must(t, m, err)
		if m.Status != "confirmed" {
			t.Fatal(m.Status)
		}
	})
	t.Run("all-entrypoints-reject-unplayable", func(t *testing.T) {
		s := NewService(factory(t))
		a := creds(t, s)
		for _, roster := range []string{"", LegacyRosterID, RegistryVersion, "unknown-roster"} {
			for _, hero := range []int{-1, 20, 126, 127, 1 << 30} {
				in := roomInput()
				in.Hero = hero
				in.RosterID = roster
				if _, err := s.CreateRoom(ctx, a.Token, in); err == nil {
					t.Fatal("room accepted", roster, hero)
				}
				if _, err := s.JoinRoom(ctx, a.Token, JoinRoom{"000001", "duel-test", hero, roster}); err == nil {
					t.Fatal("join accepted", roster, hero)
				}
				for _, pair := range [][2]int{{0, hero}, {hero, 0}} {
					if _, err := s.CreatePVE(ctx, a.Token, PVERequest{"pve_request_00001", "duel-test", pair[0], pair[1], "normal", roster}); err == nil {
						t.Fatal("PVE accepted", roster, pair)
					}
					for _, transport := range []string{"local", "broadcastchannel"} {
						if _, err := s.CreateLocal(ctx, a.Token, LocalRequest{"local_request_0001", "duel-test", pair[0], pair[1], transport, roster}); err == nil {
							t.Fatal("local accepted", roster, pair)
						}
					}
				}
			}
		}
	})
	t.Run("subset-version-snapshot-results-rematch", func(t *testing.T) {
		s := NewService(factory(t))
		s.registry = subsetRegistry(t)
		a, b := creds(t, s), creds(t, s)
		in := roomInput()
		in.Hero = 20
		in.Version = "duel-subset-test"
		in.RosterID = "test-subset-v1"
		r, err := s.CreateRoom(ctx, a.Token, in)
		r = must(t, r, err)
		for _, wrong := range []JoinRoom{{r.Code, in.Version, 0, ""}, {r.Code, in.Version, 0, LegacyRosterID}, {r.Code, "other-build", 0, in.RosterID}, {r.Code, in.Version, 21, in.RosterID}} {
			if _, err = s.JoinRoom(ctx, b.Token, wrong); err == nil {
				t.Fatal("mismatch joined")
			}
		}
		if _, err = s.JoinRoom(ctx, b.Token, JoinRoom{r.Code, in.Version, 126, in.RosterID}); err != nil {
			t.Fatal(err)
		}
		if err = s.Answer(ctx, b.Token, r.ID, in.Version, Description{"answer", "v=0\r\n"}); err != nil {
			t.Fatal(err)
		}
		m, err := s.CreateMatch(ctx, a.Token, r.ID, CreateMatch{"match_request_0001", in.Version, ""})
		m = must(t, m, err)
		if m.RosterID != in.RosterID || m.RegistryVersion != RegistryVersion || m.Players[0].Hero != 20 || m.Players[1].Hero != 126 {
			t.Fatal("snapshot mismatch", m)
		}
		if _, err = s.Ready(ctx, a.Token, m.ID, in.Version); err != nil {
			t.Fatal(err)
		}
		if _, err = s.Ready(ctx, b.Token, m.ID, in.Version); err != nil {
			t.Fatal(err)
		}
		result := completed()
		result.Version = in.Version
		if _, err = s.Submit(ctx, a.Token, m.ID, result); err != nil {
			t.Fatal(err)
		}
		m, err = s.Submit(ctx, b.Token, m.ID, result)
		m = must(t, m, err)
		if m.Status != "confirmed" || m.Submissions[0].Digest != digest(result) {
			t.Fatal("result contract changed")
		}
		next, err := s.CreateMatch(ctx, a.Token, r.ID, CreateMatch{"match_request_0002", in.Version, ""})
		next = must(t, next, err)
		if next.ID == m.ID || next.RosterID != m.RosterID || next.Players != m.Players {
			t.Fatal("rematch identity/roster changed")
		}
		for _, transport := range []string{"local", "broadcastchannel", "pve"} {
			var local Match
			if transport == "pve" {
				local, err = s.CreatePVE(ctx, a.Token, PVERequest{"pve_request_00001", in.Version, 20, 126, "hard", in.RosterID})
			} else {
				local, err = s.CreateLocal(ctx, a.Token, LocalRequest{"local_request_" + transport, in.Version, 20, 126, transport, in.RosterID})
			}
			local = must(t, local, err)
			if local.RosterID != in.RosterID || local.RegistryVersion != RegistryVersion {
				t.Fatal("single reporter snapshot lost")
			}
			local, err = s.Submit(ctx, a.Token, local.ID, result)
			local = must(t, local, err)
			if local.Status != "recorded" || local.Trust != "client_reported" {
				t.Fatal("single reporter trust changed")
			}
		}
	})
	t.Run("explicit-roster-is-part-of-idempotency", func(t *testing.T) {
		s := NewService(factory(t))
		a := creds(t, s)
		in := roomInput()
		if _, err := s.CreateRoom(ctx, a.Token, in); err != nil {
			t.Fatal(err)
		}
		in.RosterID = LegacyRosterID
		if _, err := s.CreateRoom(ctx, a.Token, in); err == nil {
			t.Fatal("changed wire body reused key")
		}
		p := PVERequest{"pve_request_00001", "duel-test", 0, 3, "normal", LegacyRosterID}
		m, err := s.CreatePVE(ctx, a.Token, p)
		m = must(t, m, err)
		same, err := s.CreatePVE(ctx, a.Token, p)
		same = must(t, same, err)
		if m.ID != same.ID {
			t.Fatal("explicit request duplicated")
		}
		p.RosterID = ""
		if _, err = s.CreatePVE(ctx, a.Token, p); err == nil {
			t.Fatal("changed roster representation accepted")
		}
	})
}

func TestRegistryMemory(t *testing.T) {
	runRegistrySuite(t, func(*testing.T) Store { return newMemory() })
}

func TestRosterHTTPContract(t *testing.T) {
	s := NewService(newMemory())
	h := &Handler{Service: s}
	a := creds(t, s)
	base := `{"requestId":"pve_request_00001","version":"duel-test","hero":0,"opponentHero":3,"aiDifficulty":"normal"`
	for i, tc := range []struct {
		tail   string
		status int
	}{
		{`}`, 201}, {`,"rosterId":"legacy-20-v1"}`, 201}, {`,"rosterId":""}`, 400}, {`,"rosterId":null}`, 400}, {`,"rosterId":20}`, 400}, {`,"rosterId":"duel-heroes-127-v1"}`, 400}, {`,"registryVersion":"duel-heroes-127-v1"}`, 400},
	} {
		body := strings.Replace(base, "pve_request_00001", fmt.Sprintf("pve_request_%05d", i), 1) + tc.tail
		req := httptest.NewRequest("POST", "/api/v1/matches/pve", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+a.Token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Fatalf("body=%s status=%d: %s", body, w.Code, w.Body)
		}
		if w.Code == 201 {
			var m map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &m)
			if m["rosterId"] != LegacyRosterID || m["registryVersion"] != RegistryVersion || m["submissions"] != nil {
				t.Fatal("bad public snapshot", m)
			}
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/registry", nil))
	var info struct {
		Heroes  []HeroIdentity   `json:"heroes"`
		Rosters []GameplayRoster `json:"gameplayRosters"`
		SHA     string           `json:"registrySha256"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &info) != nil || len(info.Heroes) != 127 || info.SHA != RegistrySHA256 || !reflect.DeepEqual(info.Rosters, s.registry.rosters) {
		t.Fatal("registry endpoint differs from validator", w.Body)
	}
}
