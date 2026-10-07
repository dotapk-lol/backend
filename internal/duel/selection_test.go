package duel

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Synthetic protocol fixture only; never activates a published runtime.
const selectionVersion = "duel-selection-test"
const selectionRoster = "selection-test-22"
const firstEpoch = "a1000000-0000-4000-8000-000000000001"
const nextEpoch = "a1000000-0000-4000-8000-000000000002"

func selectionRegistry(t *testing.T) *heroRegistry {
	t.Helper()
	var rosters []GameplayRoster
	if err := json.Unmarshal(gameplayJSON, &rosters); err != nil {
		t.Fatal(err)
	}
	rosters = append(rosters, GameplayRoster{selectionRoster, RegistryVersion, []int{1, 3, 4, 5, 7, 8, 9, 15, 17, 18, 28, 31, 32, 36, 50, 55, 57, 58, 62, 71, 81, 82}, []string{selectionVersion}})
	b, _ := json.Marshal(rosters)
	r, err := loadRegistry(registryJSON, b)
	r = must(t, r, err)
	if err = r.loadFeatures([]byte(`{"roomSelectionVersions":["duel-selection-test"]}`)); err != nil {
		t.Fatal(err)
	}
	return r
}
func selectionPair(t *testing.T, s *Service) (Credentials, Credentials, RoomView) {
	t.Helper()
	s.registry = selectionRegistry(t)
	a, b := creds(t, s), creds(t, s)
	in := roomInput()
	in.Version, in.RosterID, in.Hero = selectionVersion, selectionRoster, 1
	in.Policy = Policy{"above", 500, 250, 30, 1, 12, 10000}
	r, err := s.CreateRoom(ctx, a.Token, in)
	r = must(t, r, err)
	_, err = s.JoinRoom(ctx, b.Token, JoinRoom{r.Code, selectionVersion, 1, selectionRoster})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Answer(ctx, b.Token, r.ID, selectionVersion, Description{"answer", "v=0\r\n"}); err != nil {
		t.Fatal(err)
	}
	return a, b, r
}
func firstSelection() BeginSelection {
	return BeginSelection{selectionVersion, "begin", firstEpoch, "", ""}
}
func expectStatus(t *testing.T, err error, status int) {
	t.Helper()
	var f *Fault
	if !errors.As(err, &f) || f.Status != status {
		t.Fatalf("want %d, got %v", status, err)
	}
}
func selectBoth(t *testing.T, s *Service, a, b Credentials, r RoomView) {
	t.Helper()
	if _, err := s.BeginSelection(ctx, a.Token, r.ID, firstSelection()); err != nil {
		t.Fatal(err)
	}
	for seat, c := range []Credentials{a, b} {
		if _, err := s.LockSelection(ctx, c.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, []int{3, 82}[seat]}); err != nil {
			t.Fatal(err)
		}
	}
}
func selectedMatch(t *testing.T, s *Service, a Credentials, r RoomView) Match {
	t.Helper()
	m, err := s.CreateMatch(ctx, a.Token, r.ID, CreateMatch{"selection_match_0001", selectionVersion, firstEpoch})
	return must(t, m, err)
}
func finishSelected(t *testing.T, s *Service, a, b Credentials, m Match) {
	t.Helper()
	_, err := s.Ready(ctx, a.Token, m.ID, selectionVersion)
	if err != nil {
		t.Fatal(err)
	}
	started, err := s.Ready(ctx, b.Token, m.ID, selectionVersion)
	started = must(t, started, err)
	if started.Status != "in_progress" {
		t.Fatal(started.Status)
	}
	result := completed()
	result.Version = selectionVersion
	_, err = s.Submit(ctx, a.Token, m.ID, result)
	if err != nil {
		t.Fatal(err)
	}
	confirmed, err := s.Submit(ctx, b.Token, m.ID, result)
	confirmed = must(t, confirmed, err)
	if confirmed.Status != "confirmed" {
		t.Fatal(confirmed.Status)
	}
}

func runSelectionSuite(t *testing.T, factory func(*testing.T) Store) {
	t.Run("begin-requires-seats-answer-and-empty-first-association", func(t *testing.T) {
		s := NewService(factory(t))
		s.registry = selectionRegistry(t)
		a, b := creds(t, s), creds(t, s)
		in := roomInput()
		in.Version, in.RosterID, in.Hero = selectionVersion, selectionRoster, 1
		r, err := s.CreateRoom(ctx, a.Token, in)
		r = must(t, r, err)
		_, err = s.BeginSelection(ctx, a.Token, r.ID, firstSelection())
		expectStatus(t, err, 409)
		_, err = s.JoinRoom(ctx, b.Token, JoinRoom{r.Code, selectionVersion, 1, selectionRoster})
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.BeginSelection(ctx, a.Token, r.ID, firstSelection())
		expectStatus(t, err, 409)
		if err = s.Answer(ctx, b.Token, r.ID, selectionVersion, Description{"answer", "v=0\r\n"}); err != nil {
			t.Fatal(err)
		}
		wrong := firstSelection()
		wrong.PreviousMatchID = randomID()
		_, err = s.BeginSelection(ctx, a.Token, r.ID, wrong)
		expectStatus(t, err, 409)
		wrong = firstSelection()
		wrong.PreviousEpoch = nextEpoch
		_, err = s.BeginSelection(ctx, a.Token, r.ID, wrong)
		expectStatus(t, err, 409)
	})
	t.Run("authority-and-exact-version", func(t *testing.T) {
		s := NewService(factory(t))
		a, b, r := selectionPair(t, s)
		outsider := creds(t, s)
		for _, tc := range []struct {
			token  string
			status int
		}{{"", 401}, {outsider.Token, 403}, {b.Token, 403}} {
			_, err := s.BeginSelection(ctx, tc.token, r.ID, firstSelection())
			expectStatus(t, err, tc.status)
		}
		in := firstSelection()
		in.Version = "duel-other"
		_, err := s.BeginSelection(ctx, a.Token, r.ID, in)
		expectStatus(t, err, 409)
		in = firstSelection()
		in.Epoch = "short"
		_, err = s.BeginSelection(ctx, a.Token, r.ID, in)
		expectStatus(t, err, 400)
		_, err = s.LockSelection(ctx, outsider.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, 3})
		expectStatus(t, err, 403)
	})
	t.Run("begin-replay-does-not-reset-locks", func(t *testing.T) {
		s := NewService(factory(t))
		a, b, r := selectionPair(t, s)
		selectBoth(t, s, a, b, r)
		view, err := s.BeginSelection(ctx, a.Token, r.ID, firstSelection())
		view = must(t, view, err)
		if view.Selection.Locked != [2]bool{true, true} || view.Players[0].Hero != 3 || view.Players[1].Hero != 82 {
			t.Fatal("replay reset locks")
		}
		changed := firstSelection()
		changed.Epoch = strings.ToUpper(changed.Epoch)
		_, err = s.BeginSelection(ctx, a.Token, r.ID, changed)
		expectStatus(t, err, 409)
		changed = firstSelection()
		changed.PreviousMatchID = randomID()
		_, err = s.BeginSelection(ctx, a.Token, r.ID, changed)
		expectStatus(t, err, 409)
		m := selectedMatch(t, s, a, r)
		view, err = s.BeginSelection(ctx, a.Token, r.ID, firstSelection())
		view = must(t, view, err)
		if view.CurrentMatch != m.ID || !view.Selection.Locked[1] {
			t.Fatal("allocation broke epoch replay")
		}
		// Lost lock response can be replayed after allocation, but cannot change the hero.
		_, err = s.LockSelection(ctx, b.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, 82})
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.LockSelection(ctx, b.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, 3})
		expectStatus(t, err, 409)
	})
	t.Run("legal-roster-and-single-seat-lock", func(t *testing.T) {
		s := NewService(factory(t))
		a, b, r := selectionPair(t, s)
		_, err := s.BeginSelection(ctx, a.Token, r.ID, firstSelection())
		if err != nil {
			t.Fatal(err)
		}
		for _, hero := range []int{-1, 0, 2, 20, 126, 127} {
			_, err = s.LockSelection(ctx, a.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, hero})
			expectStatus(t, err, 400)
		}
		view, err := s.LockSelection(ctx, b.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, 82})
		view = must(t, view, err)
		if view.Selection.Locked != [2]bool{false, true} || view.Players[0].Hero != 1 || view.Players[1].Hero != 82 {
			t.Fatal("guest mutated host")
		}
		_, err = s.LockSelection(ctx, b.Token, r.ID, LockSelection{selectionVersion, "lock", nextEpoch, 82})
		expectStatus(t, err, 409)
		_, err = s.LockSelection(ctx, b.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, 3})
		expectStatus(t, err, 409)
		_, err = s.JoinRoom(ctx, b.Token, JoinRoom{r.Code, selectionVersion, 1, selectionRoster})
		if err != nil {
			t.Fatal("seat-reservation replay after selection", err)
		}
	})
	t.Run("match-lock-gates-and-idempotency", func(t *testing.T) {
		s := NewService(factory(t))
		a, b, r := selectionPair(t, s)
		in := CreateMatch{"selection_match_0001", selectionVersion, firstEpoch}
		_, err := s.CreateMatch(ctx, a.Token, r.ID, in)
		expectStatus(t, err, 409)
		_, err = s.BeginSelection(ctx, a.Token, r.ID, firstSelection())
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.LockSelection(ctx, a.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, 3})
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.CreateMatch(ctx, a.Token, r.ID, in)
		expectStatus(t, err, 409)
		_, err = s.LockSelection(ctx, b.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, 82})
		if err != nil {
			t.Fatal(err)
		}
		for _, epoch := range []string{"", nextEpoch} {
			badIn := in
			badIn.SelectionEpoch = epoch
			_, err = s.CreateMatch(ctx, a.Token, r.ID, badIn)
			expectStatus(t, err, 409)
		}
		_, err = s.CreateMatch(ctx, b.Token, r.ID, in)
		expectStatus(t, err, 403)
		m, err := s.CreateMatch(ctx, a.Token, r.ID, in)
		m = must(t, m, err)
		replay, err := s.CreateMatch(ctx, a.Token, r.ID, in)
		replay = must(t, replay, err)
		if replay.ID != m.ID {
			t.Fatal("duplicate match")
		}
		changed := in
		changed.SelectionEpoch = nextEpoch
		_, err = s.CreateMatch(ctx, a.Token, r.ID, changed)
		expectStatus(t, err, 409)
		if m.SelectionEpoch != firstEpoch || m.Players[0].Hero != 3 || m.Players[1].Hero != 82 || m.Status != "awaiting_ready" {
			t.Fatal("bad snapshot")
		}
		half, err := s.Ready(ctx, a.Token, m.ID, selectionVersion)
		half = must(t, half, err)
		if half.Status != "awaiting_ready" {
			t.Fatal("one ready started")
		}
		finishSelected(t, s, a, b, m)
		in.RequestID = "selection_match_0002"
		_, err = s.CreateMatch(ctx, a.Token, r.ID, in)
		expectStatus(t, err, 409)
	})
	t.Run("rematch-retains-room-and-immutable-evidence", func(t *testing.T) {
		s := NewService(factory(t))
		now := time.Now()
		s.Now = func() time.Time { return now }
		a, b, r := selectionPair(t, s)
		selectBoth(t, s, a, b, r)
		m := selectedMatch(t, s, a, r)
		finishSelected(t, s, a, b, m)
		old, err := s.GetMatch(ctx, a.Token, m.ID)
		old = must(t, old, err)
		snapshot := digest(old)
		now = now.Add(RoomTTL + time.Second)
		if err = s.Store.Cleanup(ctx, millis(now)); err != nil {
			t.Fatal(err)
		}
		in := BeginSelection{selectionVersion, "begin", nextEpoch, firstEpoch, m.ID}
		view, err := s.BeginSelection(ctx, a.Token, r.ID, in)
		view = must(t, view, err)
		if view.ID != r.ID || view.Players[0].ID != a.PlayerID || view.Players[1].ID != b.PlayerID || view.Players[0].Hero != 1 || view.Players[1].Hero != 1 || view.Selection.Locked != [2]bool{} {
			t.Fatal("rematch reset identity or retained locks")
		}
		for seat, c := range []Credentials{a, b} {
			_, err = s.LockSelection(ctx, c.Token, r.ID, LockSelection{selectionVersion, "lock", nextEpoch, []int{5, 7}[seat]})
			if err != nil {
				t.Fatal(err)
			}
		}
		second, err := s.CreateMatch(ctx, a.Token, r.ID, CreateMatch{"selection_match_0002", selectionVersion, nextEpoch})
		second = must(t, second, err)
		if second.ID == m.ID || second.RoomID != m.RoomID || second.Players[0].Hero != 5 || second.Players[1].Hero != 7 {
			t.Fatal("bad rematch identity")
		}
		finishSelected(t, s, a, b, second)
		old, err = s.GetMatch(ctx, a.Token, m.ID)
		old = must(t, old, err)
		if digest(old) != snapshot {
			t.Fatal("old match changed")
		}
		_, err = s.BeginSelection(ctx, a.Token, r.ID, firstSelection())
		expectStatus(t, err, 409)
		reused := BeginSelection{selectionVersion, "begin", firstEpoch, nextEpoch, second.ID}
		_, err = s.BeginSelection(ctx, a.Token, r.ID, reused)
		expectStatus(t, err, 409)
		reused.Epoch = strings.ToUpper(firstEpoch)
		_, err = s.BeginSelection(ctx, a.Token, r.ID, reused)
		expectStatus(t, err, 409)
		_, err = s.LockSelection(ctx, b.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, 82})
		expectStatus(t, err, 409)
		view, err = s.BeginSelection(ctx, a.Token, r.ID, in)
		view = must(t, view, err)
		if view.Selection.Locked != [2]bool{true, true} {
			t.Fatal("late replay erased rematch locks")
		}
	})
	t.Run("previous-match-active-pending-and-expired", func(t *testing.T) {
		s := NewService(factory(t))
		now := time.Now()
		s.Now = func() time.Time { return now }
		a, b, r := selectionPair(t, s)
		selectBoth(t, s, a, b, r)
		m := selectedMatch(t, s, a, r)
		in := BeginSelection{selectionVersion, "begin", nextEpoch, firstEpoch, m.ID}
		_, err := s.BeginSelection(ctx, a.Token, r.ID, in)
		expectStatus(t, err, 409)
		_, err = s.Ready(ctx, a.Token, m.ID, selectionVersion)
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.Ready(ctx, b.Token, m.ID, selectionVersion)
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.BeginSelection(ctx, a.Token, r.ID, in)
		expectStatus(t, err, 409)
		result := completed()
		result.Version = selectionVersion
		_, err = s.Submit(ctx, a.Token, m.ID, result)
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.BeginSelection(ctx, a.Token, r.ID, in)
		expectStatus(t, err, 409)
		wrong := in
		wrong.PreviousMatchID = randomID()
		_, err = s.BeginSelection(ctx, a.Token, r.ID, wrong)
		expectStatus(t, err, 409)
		wrong = in
		wrong.PreviousEpoch = nextEpoch
		_, err = s.BeginSelection(ctx, a.Token, r.ID, wrong)
		expectStatus(t, err, 409)
		now = now.Add(ReportTTL + time.Second)
		_, err = s.BeginSelection(ctx, a.Token, r.ID, in)
		if err != nil {
			t.Fatal(err)
		}
		old, err := s.GetMatch(ctx, a.Token, m.ID)
		old = must(t, old, err)
		if old.Status != "aborted" {
			t.Fatal(old.Status)
		}
	})
	t.Run("concurrent-locks-and-allocation", func(t *testing.T) {
		s := NewService(factory(t))
		a, b, r := selectionPair(t, s)
		_, err := s.BeginSelection(ctx, a.Token, r.ID, firstSelection())
		if err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		for seat, c := range []Credentials{a, b} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, e := s.LockSelection(ctx, c.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, []int{3, 82}[seat]})
				errs <- e
			}()
		}
		wg.Wait()
		close(errs)
		for e := range errs {
			if e != nil {
				t.Fatal(e)
			}
		}
		view, err := s.GetRoom(ctx, a.Token, r.ID)
		view = must(t, view, err)
		if view.Selection.Locked != [2]bool{true, true} {
			t.Fatal("lost concurrent lock")
		}
		results := make(chan error, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, e := s.CreateMatch(ctx, a.Token, r.ID, CreateMatch{fmt.Sprintf("concurrent_match_%02d", i), selectionVersion, firstEpoch})
				results <- e
			}()
		}
		wg.Wait()
		close(results)
		wins := 0
		for e := range results {
			if e == nil {
				wins++
			} else {
				expectStatus(t, e, 409)
			}
		}
		if wins != 1 {
			t.Fatal("allocated", wins)
		}
	})
	t.Run("closed-and-participant-expiry", func(t *testing.T) {
		s := NewService(factory(t))
		a, b, r := selectionPair(t, s)
		selectBoth(t, s, a, b, r)
		err := s.Store.Run(ctx, func(tx Tx) error {
			return tx.Put("session", hash(b.Token), Session{b.PlayerID, millis(s.Now()) - 1}, 0)
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.BeginSelection(ctx, a.Token, r.ID, firstSelection())
		expectStatus(t, err, 409)
		_, err = s.LockSelection(ctx, a.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, 3})
		expectStatus(t, err, 409)
		_, err = s.CreateMatch(ctx, a.Token, r.ID, CreateMatch{"selection_match_0001", selectionVersion, firstEpoch})
		expectStatus(t, err, 409)
		err = s.Store.Run(ctx, func(tx Tx) error { return tx.Put("session", hash(b.Token), Session{b.PlayerID, b.Expires}, b.Expires) })
		if err != nil {
			t.Fatal(err)
		}
		if err = s.CloseRoom(ctx, a.Token, r.ID); err != nil {
			t.Fatal(err)
		}
		_, err = s.BeginSelection(ctx, a.Token, r.ID, firstSelection())
		expectStatus(t, err, 409)
		_, err = s.LockSelection(ctx, b.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, 82})
		expectStatus(t, err, 409)
	})
	t.Run("room-24h-expiry-with-fresh-sessions", func(t *testing.T) {
		s := NewService(factory(t))
		a, b, r := selectionPair(t, s)
		selectBoth(t, s, a, b, r)
		err := s.Store.Run(ctx, func(tx Tx) error {
			var room Room
			if e := tx.Get("room", r.ID, &room); e != nil {
				return e
			}
			room.CreatedAt = millis(s.Now()) - SessionTTL.Milliseconds()
			return tx.Put("room", room.ID, room, room.Expires)
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.BeginSelection(ctx, a.Token, r.ID, firstSelection())
		expectStatus(t, err, 409)
		_, err = s.LockSelection(ctx, b.Token, r.ID, LockSelection{selectionVersion, "lock", firstEpoch, 82})
		expectStatus(t, err, 409)
		_, err = s.CreateMatch(ctx, a.Token, r.ID, CreateMatch{"selection_match_0001", selectionVersion, firstEpoch})
		expectStatus(t, err, 409)
	})
}

func TestSelectionMemory(t *testing.T) {
	runSelectionSuite(t, func(*testing.T) Store { return newMemory() })
}

func TestSelectionFeaturesAndPolicy(t *testing.T) {
	r := selectionRegistry(t)
	for _, raw := range []string{`{"roomSelectionVersions":["unknown"]}`, `{"roomSelectionVersions":["duel-selection-test","duel-selection-test"]}`} {
		if err := r.loadFeatures([]byte(raw)); err == nil {
			t.Fatal("invalid binding accepted")
		}
	}
	s := NewService(newMemory())
	a := creds(t, s)
	in := roomInput()
	in.Policy = Policy{"above", 500, 250, 30, 1, 12, 10000}
	_, err := s.CreateRoom(ctx, a.Token, in)
	expectStatus(t, err, 400)
	s.registry = selectionRegistry(t)
	in.Version, in.RosterID, in.Hero = selectionVersion, selectionRoster, 1
	for _, p := range []Policy{{"below", 500, 250, 30, 1, 12, 10000}, {"above", 501, 250, 30, 1, 12, 10000}, {"above", 500, 250, 30, 0, 12, 10000}} {
		in.Policy = p
		_, err = s.CreateRoom(ctx, a.Token, in)
		expectStatus(t, err, 400)
	}
	in.Policy = Policy{"above", 500, 250, 30, 1, 12, 10000}
	view, err := s.CreateRoom(ctx, a.Token, in)
	view = must(t, view, err)
	if view.Policy != in.Policy {
		t.Fatal("casual tuple changed")
	}
	// Every published fixture member is legal; catalog-only and paused members are not.
	for _, hero := range s.registry.byRoster[selectionRoster].HeroIDs {
		if _, err = s.registry.resolve(selectionRoster, selectionVersion, hero); err != nil {
			t.Fatal(hero, err)
		}
	}
	legacy := NewService(newMemory())
	host, _, room := pair(t, legacy)
	_, err = legacy.BeginSelection(ctx, host.Token, room.ID, BeginSelection{"duel-test", "begin", firstEpoch, "", ""})
	expectStatus(t, err, 409)
	_, err = legacy.CreateMatch(ctx, host.Token, room.ID, CreateMatch{"legacy_match_test1", "duel-test", firstEpoch})
	expectStatus(t, err, 409)
}

func TestSelectionHTTPShapeAndPrivacy(t *testing.T) {
	s := NewService(newMemory())
	a, b, r := selectionPair(t, s)
	h := Handler{Service: s, Origin: "https://example.invalid"}
	call := func(token, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/api/v1/rooms/"+r.ID+"/selection", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w
	}
	beg, _ := json.Marshal(firstSelection())
	for _, body := range []string{`{"version":"duel-selection-test","action":"begin","epoch":"` + firstEpoch + `"}`, string(beg[:len(beg)-1]) + `,"hero":3}`, `{"version":"duel-selection-test","action":"lock","epoch":"` + firstEpoch + `","hero":3,"seat":1}`, `{"version":"duel-selection-test","action":"lock","epoch":"` + firstEpoch + `","hero":3,"opponentHero":82}`, `{"version":"duel-selection-test","action":"lock","epoch":"` + firstEpoch + `","hero":null}`} {
		if w := call(a.Token, body); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if w := call("", string(beg)); w.Code != 401 {
		t.Fatal(w.Code)
	}
	if w := call(a.Token, string(beg)); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	lock, _ := json.Marshal(LockSelection{selectionVersion, "lock", firstEpoch, 82})
	w := call(b.Token, string(lock))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var view map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	sel := view["selection"].(map[string]any)
	if len(sel) != 4 || view["tokens"] != nil || sel["matchId"] != nil || strings.Contains(w.Body.String(), a.Token) {
		t.Fatal("private selection/auth state leaked")
	}
	for _, epoch := range []string{`""`, `null`, `123`} {
		req := httptest.NewRequest("POST", "/api/v1/rooms/"+r.ID+"/matches", strings.NewReader(`{"requestId":"selection_http_01","version":"duel-selection-test","selectionEpoch":`+epoch+`}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+a.Token)
		w = httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatal("optional epoch shape", w.Code)
		}
	}
}
