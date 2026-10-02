package duel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type memory struct {
	mu   sync.Mutex
	rows map[string]json.RawMessage
}
type memoryTx map[string]json.RawMessage

func newMemory() *memory { return &memory{rows: map[string]json.RawMessage{}} }
func (m *memory) Run(_ context.Context, fn func(Tx) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	copy := memoryTx{}
	for k, v := range m.rows {
		copy[k] = v
	}
	if e := fn(copy); e != nil {
		return e
	}
	m.rows = copy
	return nil
}
func (m *memory) Ping(context.Context) error           { return nil }
func (m *memory) Cleanup(context.Context, int64) error { return nil }
func (m memoryTx) Get(k, id string, v any) error {
	b, ok := m[k+":"+id]
	if !ok {
		return ErrMissing
	}
	return json.Unmarshal(b, v)
}
func (m memoryTx) Put(k, id string, v any, _ int64) error {
	b, e := json.Marshal(v)
	m[k+":"+id] = b
	return e
}
func (m memoryTx) Insert(k, id string, v any, x int64) error {
	if _, ok := m[k+":"+id]; ok {
		return ErrExists
	}
	return m.Put(k, id, v, x)
}
func (m memoryTx) Delete(k, id string) error { delete(m, k+":"+id); return nil }

var ctx = context.Background()

func must[T any](t *testing.T, v T, e error) T {
	t.Helper()
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func creds(t *testing.T, s *Service) Credentials {
	t.Helper()
	v, e := s.NewSession(ctx)
	return must(t, v, e)
}
func policy() Policy { return Policy{"above", 200, 30, 5, 24, 30, 3000} }
func roomInput() CreateRoom {
	return CreateRoom{"request_1234567890", "duel-test", 0, Description{"offer", "v=0\r\n"}, policy()}
}
func completed() Result {
	return Result{"duel-test", "completed", []Round{{1, 0, 1200}, {2, -1, 0}, {3, 0, 2000}}, [2]int{2, 0}, 0, ""}
}
func pair(t *testing.T, s *Service) (Credentials, Credentials, RoomView) {
	a, b := creds(t, s), creds(t, s)
	r, e := s.CreateRoom(ctx, a.Token, roomInput())
	r = must(t, r, e)
	_, e = s.JoinRoom(ctx, b.Token, JoinRoom{r.Code, "duel-test", 3})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Answer(ctx, b.Token, r.ID, "duel-test", Description{"answer", "v=0\r\n"}); e != nil {
		t.Fatal(e)
	}
	return a, b, r
}
func start(t *testing.T, s *Service, a, b Credentials, r RoomView, key string) Match {
	t.Helper()
	m, e := s.CreateMatch(ctx, a.Token, r.ID, CreateMatch{key, "duel-test"})
	m = must(t, m, e)
	_, e = s.Ready(ctx, a.Token, m.ID, "duel-test")
	if e != nil {
		t.Fatal(e)
	}
	m, e = s.Ready(ctx, b.Token, m.ID, "duel-test")
	return must(t, m, e)
}
func runSuite(t *testing.T, factory func(*testing.T) Store) {
	t.Run("code-leading-zero-collision-expiry", func(t *testing.T) {
		s := NewService(factory(t))
		now := time.Now()
		s.Now = func() time.Time { return now }
		s.Code = func() (string, error) { return "000007", nil }
		a := creds(t, s)
		r, e := s.CreateRoom(ctx, a.Token, roomInput())
		r = must(t, r, e)
		if r.Code != "000007" {
			t.Fatal(r.Code)
		}
		b := creds(t, s)
		if _, e = s.CreateRoom(ctx, b.Token, roomInput()); e == nil {
			t.Fatal("collision must not overwrite")
		}
		now = now.Add(RoomTTL + time.Second)
		r2, e := s.CreateRoom(ctx, b.Token, roomInput())
		r2 = must(t, r2, e)
		if r2.ID == r.ID {
			t.Fatal("recycled code reused room identity")
		}
		if e = s.CloseRoom(ctx, a.Token, r.ID); e != nil {
			t.Fatal(e)
		}
		c := creds(t, s)
		if _, e = s.JoinRoom(ctx, c.Token, JoinRoom{r2.Code, "duel-test", 1}); e != nil {
			t.Fatal("old room close deleted recycled code", e)
		}
	})
	t.Run("concurrent-guest-single-seat", func(t *testing.T) {
		s := NewService(factory(t))
		a := creds(t, s)
		r, e := s.CreateRoom(ctx, a.Token, roomInput())
		r = must(t, r, e)
		var wg sync.WaitGroup
		var mu sync.Mutex
		wins := 0
		for i := 0; i < 12; i++ {
			b := creds(t, s)
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, e := s.JoinRoom(ctx, b.Token, JoinRoom{r.Code, "duel-test", 3})
				if e == nil {
					mu.Lock()
					wins++
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
		if wins != 1 {
			t.Fatalf("claimed seats=%d", wins)
		}
	})
	t.Run("room-idempotency-and-identity", func(t *testing.T) {
		s := NewService(factory(t))
		a := creds(t, s)
		r, e := s.CreateRoom(ctx, a.Token, roomInput())
		r = must(t, r, e)
		r2, e := s.CreateRoom(ctx, a.Token, roomInput())
		r2 = must(t, r2, e)
		if r.ID != r2.ID {
			t.Fatal("duplicate room")
		}
		in := roomInput()
		in.Hero = 1
		if _, e = s.CreateRoom(ctx, a.Token, in); e == nil {
			t.Fatal("request body conflict accepted")
		}
		if _, e = s.JoinRoom(ctx, a.Token, JoinRoom{r.Code, "duel-test", 0}); e == nil {
			t.Fatal("self-join accepted")
		}
		b := creds(t, s)
		if _, e = s.GetRoom(ctx, b.Token, r.ID); e == nil {
			t.Fatal("outsider read room")
		}
		if _, e = s.JoinRoom(ctx, b.Token, JoinRoom{r.Code, "other", 0}); e == nil {
			t.Fatal("wrong version")
		}
	})
	t.Run("confirmation-idempotency-and-rematch", func(t *testing.T) {
		s := NewService(factory(t))
		a, b, r := pair(t, s)
		m := start(t, s, a, b, r, "match_request_0001")
		if m.Status != "in_progress" || m.StartedAt == 0 {
			t.Fatal(m)
		}
		if _, e := s.CreateMatch(ctx, a.Token, r.ID, CreateMatch{"match_request_0002", "duel-test"}); e == nil {
			t.Fatal("overlapping match")
		}
		v, e := s.Submit(ctx, a.Token, m.ID, completed())
		v = must(t, v, e)
		if v.Status != "pending" || v.Winner != -1 {
			t.Fatal("single client awarded win")
		}
		v, e = s.Submit(ctx, b.Token, m.ID, completed())
		v = must(t, v, e)
		if v.Status != "confirmed" || v.Score != [2]int{2, 0} {
			t.Fatal(v)
		}
		var wg sync.WaitGroup
		for i := 0; i < 12; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				x, e := s.Submit(ctx, a.Token, m.ID, completed())
				if e != nil || x.Status != "confirmed" {
					t.Error(e, x.Status)
				}
			}()
		}
		wg.Wait()
		other := completed()
		other.Rounds = []Round{{1, 1, 0}, {2, 1, 0}}
		other.Score = [2]int{0, 2}
		other.Winner = 1
		if _, e = s.Submit(ctx, a.Token, m.ID, other); e == nil {
			t.Fatal("mutated report accepted")
		}
		m2 := start(t, s, a, b, r, "match_request_0002")
		if m.ID == m2.ID {
			t.Fatal("rematch reused ID")
		}
	})
	t.Run("dispute-and-outsider", func(t *testing.T) {
		s := NewService(factory(t))
		a, b, r := pair(t, s)
		m := start(t, s, a, b, r, "match_request_0001")
		c := creds(t, s)
		if _, e := s.Submit(ctx, c.Token, m.ID, completed()); e == nil {
			t.Fatal("outsider report")
		}
		_, e := s.Submit(ctx, a.Token, m.ID, completed())
		if e != nil {
			t.Fatal(e)
		}
		other := completed()
		other.Rounds = []Round{{1, 1, 0}, {2, 1, 0}}
		other.Score = [2]int{0, 2}
		other.Winner = 1
		v, e := s.Submit(ctx, b.Token, m.ID, other)
		v = must(t, v, e)
		if v.Status != "disputed" || v.Winner != -1 {
			t.Fatal(v)
		}
	})
	t.Run("pending-timeout-and-late-report", func(t *testing.T) {
		s := NewService(factory(t))
		now := time.Now()
		s.Now = func() time.Time { return now }
		a, b, r := pair(t, s)
		m := start(t, s, a, b, r, "match_request_0001")
		_, e := s.Submit(ctx, a.Token, m.ID, completed())
		if e != nil {
			t.Fatal(e)
		}
		now = now.Add(ReportTTL + time.Second)
		v, e := s.Submit(ctx, b.Token, m.ID, completed())
		v = must(t, v, e)
		if v.Status != "aborted" || v.Reason != "result_timeout" || v.Winner != -1 || v.Submissions[1] != nil {
			t.Fatal(v)
		}
		_ = start(t, s, a, b, r, "match_request_0002")
	})
	t.Run("no-report-timeout", func(t *testing.T) {
		s := NewService(factory(t))
		now := time.Now()
		s.Now = func() time.Time { return now }
		a, b, r := pair(t, s)
		m := start(t, s, a, b, r, "match_request_0001")
		now = now.Add(MatchTTL + time.Second)
		m, e := s.GetMatch(ctx, a.Token, m.ID)
		m = must(t, m, e)
		if m.Status != "aborted" || m.Winner != -1 {
			t.Fatal(m)
		}
	})
	t.Run("pve-no-ai-credential", func(t *testing.T) {
		s := NewService(factory(t))
		a := creds(t, s)
		in := PVERequest{"pve_request_00001", "duel-test", 0, 3, "normal"}
		m, e := s.CreatePVE(ctx, a.Token, in)
		m = must(t, m, e)
		same, e := s.CreatePVE(ctx, a.Token, in)
		same = must(t, same, e)
		if same.ID != m.ID {
			t.Fatal("duplicate PVE")
		}
		m, e = s.Submit(ctx, a.Token, m.ID, completed())
		m = must(t, m, e)
		if m.Status != "recorded" || m.Trust != "client_reported" || m.Players[1].ID != "ai" {
			t.Fatal(m)
		}
		c := creds(t, s)
		if _, e = s.GetMatch(ctx, c.Token, m.ID); e == nil {
			t.Fatal("outsider pve access")
		}
	})
	t.Run("abort-never-win", func(t *testing.T) {
		s := NewService(factory(t))
		a, b, r := pair(t, s)
		m := start(t, s, a, b, r, "match_request_0001")
		v := Result{Version: "duel-test", Outcome: "aborted", Rounds: []Round{}, Winner: -1, Reason: "disconnect"}
		_, e := s.Submit(ctx, a.Token, m.ID, v)
		if e != nil {
			t.Fatal(e)
		}
		m, e = s.Submit(ctx, b.Token, m.ID, v)
		m = must(t, m, e)
		if m.Status != "aborted" || m.Winner != -1 {
			t.Fatal(m)
		}
	})
	t.Run("invalid-structure", func(t *testing.T) {
		s := NewService(factory(t))
		a, b, r := pair(t, s)
		m := start(t, s, a, b, r, "match_request_0001")
		cases := []Result{completed(), completed(), completed(), completed(), completed()}
		cases[0].Score = [2]int{2, 1}
		cases[1].Winner = 1
		cases[2].Rounds = append(cases[2].Rounds, Round{4, 0, 0})
		cases[3].Version = "old"
		cases[4].Rounds[0].Number = 2
		for i, v := range cases {
			if _, e := s.Submit(ctx, a.Token, m.ID, v); e == nil {
				t.Errorf("invalid case %d", i)
			}
		}
	})
	t.Run("rate-concurrency", func(t *testing.T) {
		s := NewService(factory(t))
		var wg sync.WaitGroup
		var mu sync.Mutex
		wins := 0
		for i := 0; i < 30; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if s.Limit(ctx, "127.0.0.1", "join", 10, 60) == nil {
					mu.Lock()
					wins++
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
		if wins != 10 {
			t.Fatalf("accepted %d want 10", wins)
		}
	})
	t.Run("expired-auth", func(t *testing.T) {
		s := NewService(factory(t))
		now := time.Now()
		s.Now = func() time.Time { return now }
		a := creds(t, s)
		now = now.Add(SessionTTL + time.Second)
		if _, e := s.CreateRoom(ctx, a.Token, roomInput()); e == nil {
			t.Fatal("expired session accepted")
		}
	})
}
func TestMemorySuite(t *testing.T) { runSuite(t, func(*testing.T) Store { return newMemory() }) }
func TestHTTPBoundaries(t *testing.T) {
	s := NewService(newMemory())
	h := &Handler{Service: s, Origin: "https://example.test"}
	a := creds(t, s)
	valid, _ := json.Marshal(PVERequest{"pve_request_00001", "duel-test", 0, 3, "normal"})
	for _, tc := range []struct {
		name, body, origin string
		want               int
	}{{"ok", string(valid), "https://example.test", 201}, {"origin", string(valid), "https://evil.test", 403}, {"missing", `{"requestId":"pve_request_00001"}`, "", 400}, {"null", `null`, "", 400}, {"unknown", strings.TrimSuffix(string(valid), "}") + `,"extra":1}`, "", 400}, {"large", strings.Repeat("x", 45001), "", 413}, {"trailing", string(valid) + `{}`, "", 400}} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/api/v1/matches/pve", strings.NewReader(tc.body))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Authorization", "Bearer "+a.Token)
			r.Header.Set("Origin", tc.origin)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("%d %s", w.Code, w.Body)
			}
		})
	}
	for _, badScore := range []string{`[2]`, `[2,0,1]`, `null`} {
		r := completed()
		b, _ := json.Marshal(r)
		raw := strings.Replace(string(b), `[2,0]`, badScore, 1)
		req := httptest.NewRequest("POST", "/", strings.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		if e := decode(httptest.NewRecorder(), req, &r); e == nil {
			t.Fatal("invalid score accepted", badScore)
		}
	}
}
func TestNoResultLeak(t *testing.T) {
	m := Match{Submissions: [2]*Submission{{Digest: "secret-report", Result: completed()}, nil}}
	b, _ := json.Marshal(publicMatch(m))
	if strings.Contains(string(b), "secret-report") || strings.Contains(string(b), "rounds") {
		t.Fatal("uncommitted peer report leaked")
	}
}
func TestRandomCodes(t *testing.T) {
	for i := 0; i < 1000; i++ {
		c, e := randomCode()
		if e != nil || !codePattern.MatchString(c) {
			t.Fatal(c, e)
		}
	}
}
func TestStoreRollback(t *testing.T) {
	m := newMemory()
	_ = m.Run(ctx, func(t Tx) error { _ = t.Insert("code", "000000", 1, 1); return fmt.Errorf("rollback") })
	e := m.Run(ctx, func(t Tx) error { var n int; return t.Get("code", "000000", &n) })
	if !errors.Is(e, ErrMissing) {
		t.Fatal("transaction did not roll back")
	}
}
