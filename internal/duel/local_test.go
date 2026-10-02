package duel

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func localInput(transport string) LocalRequest {
	return LocalRequest{"local_request_0001", "duel-test", 1, 4, transport, ""}
}
func runLocalSuite(t *testing.T, factory func(*testing.T) Store) {
	for _, transport := range []string{"local", "broadcastchannel"} {
		t.Run("local-pvp-"+transport, func(t *testing.T) {
			s := NewService(factory(t))
			a := creds(t, s)
			in := localInput(transport)
			m, e := s.CreateLocal(ctx, a.Token, in)
			m = must(t, m, e)
			if m.Mode != "pvp" || m.Transport != transport || m.Trust != "client_reported" || m.Status != "in_progress" || m.ReporterPlayerID != a.PlayerID || m.ParticipantKinds != [2]string{"local_slot", "local_slot"} || m.AIDifficulty != "" {
				t.Fatal("local identity semantics", m)
			}
			if m.Players[0].ID == a.PlayerID || m.Players[1].ID == a.PlayerID || m.Players[0].ID == m.Players[1].ID || m.Players[1].ID == "ai" {
				t.Fatal("slots are not independent anonymous local IDs")
			}
			var wg sync.WaitGroup
			for i := 0; i < 12; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					retry, e := s.CreateLocal(ctx, a.Token, in)
					if e != nil || retry.ID != m.ID {
						t.Error("idempotent create", e)
					}
				}()
			}
			wg.Wait()
			outsider := creds(t, s)
			if _, e = s.Submit(ctx, outsider.Token, m.ID, completed()); e == nil {
				t.Fatal("outsider recorded local match")
			}
			changed := in
			changed.Hero = 3
			if _, e = s.CreateLocal(ctx, a.Token, changed); e == nil {
				t.Fatal("changed idempotency payload accepted")
			}
			m, e = s.Submit(ctx, a.Token, m.ID, completed())
			m = must(t, m, e)
			if m.Status != "recorded" || m.Submissions[1] != nil || m.Winner != 0 {
				t.Fatal("local must never become peer confirmed", m)
			}
			retry, e := s.Submit(ctx, a.Token, m.ID, completed())
			retry = must(t, retry, e)
			if retry.ID != m.ID || retry.Status != "recorded" {
				t.Fatal(retry)
			}
			changed = in
			changed.RequestID = "local_request_0002"
			next, e := s.CreateLocal(ctx, a.Token, changed)
			next = must(t, next, e)
			if next.ID == m.ID || next.Players[0].ID == m.Players[0].ID {
				t.Fatal("local rematch reused identity")
			}
		})
	}
	t.Run("local-pvp-abort-and-timeout", func(t *testing.T) {
		s := NewService(factory(t))
		now := time.Now()
		s.Now = func() time.Time { return now }
		a := creds(t, s)
		m, e := s.CreateLocal(ctx, a.Token, localInput("local"))
		m = must(t, m, e)
		m, e = s.Submit(ctx, a.Token, m.ID, Result{Version: "duel-test", Outcome: "aborted", Rounds: []Round{}, Winner: -1, Reason: "left"})
		m = must(t, m, e)
		if m.Status != "aborted" || m.Winner != -1 || m.Reason != "left" {
			t.Fatal(m)
		}
		in := localInput("broadcastchannel")
		in.RequestID = "local_request_0002"
		m, e = s.CreateLocal(ctx, a.Token, in)
		m = must(t, m, e)
		now = now.Add(MatchTTL + time.Second)
		m, e = s.GetMatch(ctx, a.Token, m.ID)
		m = must(t, m, e)
		if m.Status != "aborted" || m.Reason != "result_timeout" {
			t.Fatal(m)
		}
	})
	t.Run("local-pvp-invalid-transport-and-heroes", func(t *testing.T) {
		s := NewService(factory(t))
		a := creds(t, s)
		for _, v := range []LocalRequest{{"local_request_0001", "duel-test", 0, 1, "webrtc", ""}, {"local_request_0001", "duel-test", 0, 20, "local", ""}, {"local_request_0001", "duel-test", -1, 1, "broadcastchannel", ""}} {
			if _, e := s.CreateLocal(ctx, a.Token, v); e == nil {
				t.Fatal("invalid local request accepted")
			}
		}
	})
}
func TestLocalMemorySuite(t *testing.T) {
	runLocalSuite(t, func(*testing.T) Store { return newMemory() })
}
func TestLocalHTTP(t *testing.T) {
	s := NewService(newMemory())
	h := &Handler{Service: s}
	a := creds(t, s)
	valid, _ := json.Marshal(localInput("local"))
	for _, tc := range []struct {
		body string
		want int
	}{{string(valid), 201}, {strings.TrimSuffix(string(valid), "}") + `,"trust":"peer_agreement"}`, 400}, {`{"requestId":"local_request_0002","version":"duel-test","hero":0,"opponentHero":1}`, 400}} {
		req := httptest.NewRequest("POST", "/api/v1/matches/local", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+a.Token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}
