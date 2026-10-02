package duel

import (
	"reflect"
	"testing"
	"time"
)

func abortResult(reason string, rounds []Round) Result {
	v := Result{Version: "duel-test", Outcome: "aborted", Rounds: rounds, Winner: -1, Reason: reason}
	if v.Rounds == nil {
		v.Rounds = []Round{}
	}
	for _, r := range rounds {
		if r.Winner >= 0 {
			v.Score[r.Winner]++
		}
	}
	return v
}
func runReconciliationSuite(t *testing.T, factory func(*testing.T) Store) {
	complete0 := completed()
	complete1 := completed()
	complete1.Rounds = []Round{{1, 1, 2000}, {2, 1, 1000}}
	complete1.Score = [2]int{0, 2}
	complete1.Winner = 1
	changedTime := completed()
	changedTime.Rounds = append([]Round(nil), changedTime.Rounds...)
	changedTime.Rounds[0].RemainingMS++
	cases := []struct {
		name                      string
		a, b                      Result
		status, reason, agreement string
		score                     [2]int
		winner                    int
	}{
		{"both-aborted-same", abortResult("left", nil), abortResult("left", nil), "aborted", "left", "peer_agreed", [2]int{}, -1},
		{"left-disconnect", abortResult("left", nil), abortResult("disconnect", nil), "aborted", "interrupted", "peer_agreed", [2]int{}, -1},
		{"partial-round-views", abortResult("left", []Round{{1, 0, 1000}}), abortResult("disconnect", []Round{{1, 0, 500}}), "aborted", "interrupted", "peer_agreed", [2]int{1, 0}, -1},
		{"different-partial-scores", abortResult("left", []Round{{1, 0, 0}}), abortResult("disconnect", nil), "aborted", "interrupted", "unresolved", [2]int{}, -1},
		{"completed-vs-aborted", complete0, abortResult("disconnect", nil), "disputed", "outcome_conflict", "unresolved", [2]int{}, -1},
		{"completed-winner-conflict", complete0, complete1, "disputed", "conflicting_reports", "unresolved", [2]int{}, -1},
		{"completed-round-conflict", complete0, changedTime, "disputed", "conflicting_reports", "unresolved", [2]int{}, -1},
		{"completed-agree", complete0, complete0, "confirmed", "", "peer_agreed", [2]int{2, 0}, 0},
	}
	for _, tc := range cases {
		for _, reverse := range []bool{false, true} {
			name := tc.name
			if reverse {
				name += "-reverse-arrival"
			}
			t.Run(name, func(t *testing.T) {
				s := NewService(factory(t))
				now := time.Now()
				s.Now = func() time.Time { return now }
				a, b, r := pair(t, s)
				m := start(t, s, a, b, r, "reconcile_match_001")
				first, second := a, b
				firstReport, secondReport := tc.a, tc.b
				if reverse {
					first, second = b, a
					firstReport, secondReport = tc.b, tc.a
				}
				p, e := s.Submit(ctx, first.Token, m.ID, firstReport)
				p = must(t, p, e)
				if p.Status != "pending" {
					t.Fatal(p.Status)
				}
				deadline := p.Deadline
				now = now.Add(time.Second)
				replay, e := s.Submit(ctx, first.Token, m.ID, firstReport)
				replay = must(t, replay, e)
				if replay.Deadline != deadline || !reflect.DeepEqual(replay.Submissions, p.Submissions) {
					t.Fatal("pending retry altered report/deadline")
				}
				final, e := s.Submit(ctx, second.Token, m.ID, secondReport)
				final = must(t, final, e)
				if final.Status != tc.status || final.Reason != tc.reason || final.ScoreAgreement != tc.agreement || final.Score != tc.score || final.Winner != tc.winner {
					t.Fatalf("resolution got status=%s reason=%s agreement=%s score=%v winner=%d", final.Status, final.Reason, final.ScoreAgreement, final.Score, final.Winner)
				}
				if sqlStore, ok := s.Store.(*SQLStore); ok {
					var appearances int
					if err := sqlStore.DB.QueryRow("SELECT COALESCE(SUM(appearances),0) FROM duel_hero_balance_v2").Scan(&appearances); err != nil {
						t.Fatal(err)
					}
					want := 0
					if tc.status == "confirmed" {
						want = 2
					}
					if appearances != want {
						t.Fatal("aborted/disputed leaked into win rate", appearances, want)
					}
				}
				for i, want := range []Result{tc.a, tc.b} {
					if final.Submissions[i].Digest != digest(want) || !reflect.DeepEqual(final.Submissions[i].Result, want) {
						t.Fatal("original report changed", i)
					}
				}
				now = now.Add(MatchTTL)
				again, e := s.Submit(ctx, first.Token, m.ID, firstReport)
				again = must(t, again, e)
				if !reflect.DeepEqual(again, final) {
					t.Fatal("terminal retry changed result")
				}
				changed := firstReport
				changed.Version = "other"
				if _, e = s.Submit(ctx, first.Token, m.ID, changed); e == nil {
					t.Fatal("mutated retry accepted")
				}
			})
		}
	}
	t.Run("aborted-report-timeout-is-immutable", func(t *testing.T) {
		s := NewService(factory(t))
		now := time.Now()
		s.Now = func() time.Time { return now }
		a, b, r := pair(t, s)
		m := start(t, s, a, b, r, "reconcile_match_001")
		original := abortResult("left", []Round{{1, 0, 1000}})
		p, e := s.Submit(ctx, a.Token, m.ID, original)
		p = must(t, p, e)
		now = now.Add(ReportTTL + time.Second)
		expired, e := s.Submit(ctx, b.Token, m.ID, abortResult("disconnect", nil))
		expired = must(t, expired, e)
		if expired.Status != "aborted" || expired.Reason != "result_timeout" || expired.Submissions[1] != nil || !reflect.DeepEqual(expired.Submissions[0], p.Submissions[0]) {
			t.Fatal("late report altered expired evidence")
		}
		replay, e := s.Submit(ctx, a.Token, m.ID, original)
		replay = must(t, replay, e)
		if !reflect.DeepEqual(replay, expired) {
			t.Fatal("expired retry altered terminal result")
		}
	})
}
func TestReconciliationMemory(t *testing.T) {
	runReconciliationSuite(t, func(*testing.T) Store { return newMemory() })
}
