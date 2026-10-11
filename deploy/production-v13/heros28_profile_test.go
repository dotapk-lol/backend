package duel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"
)

// Exact final six-hero fingerprint; both manifests and SQL must match.
const heros28Version = "duel-1518053e640d9203bbf2"
const heros28RetainedVersion = "duel-6750ae2e1cfd1c3cdd42"
const heros28Roster = "arena-heros28-v1"

func TestHeros28ProductionBoundary(t *testing.T) {
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 14, 15, 17, 18, 19, 28, 31, 32, 36, 50, 55, 57, 58, 62, 71, 81, 82}
	s := NewService(newMemory())
	r := s.registry.byRoster[heros28Roster]
	if r.RegistryVersion != RegistryVersion || !reflect.DeepEqual(r.HeroIDs, ids) || !reflect.DeepEqual(r.GameVersions, []string{heros28RetainedVersion, heros28Version}) || !s.registry.selectionEnabled(heros28Version) || !s.registry.selectionEnabled(heros28RetainedVersion) || s.registry.selectionEnabled("duel-56ab6091b8eb3f5603ec") {
		t.Fatal("28 profile fingerprint drift")
	}
	// Exact old membership/builds are independently pinned by the22 fixture.
	old46 := s.registry.byRoster["arena-first22-46-v1"]
	if !reflect.DeepEqual(old46.HeroIDs, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 25, 31, 45, 100, 21, 28, 29, 32, 36, 42, 47, 50, 55, 57, 58, 62, 71, 81, 82, 94, 99, 104, 106, 117, 121, 124}) || !reflect.DeepEqual(old46.GameVersions, []string{"duel-e63dafb5ae2070a90f8b", "duel-7e767a8ed2b995465875"}) || len(s.registry.byRoster[LegacyRosterID].HeroIDs) != 20 {
		t.Fatal("historical bindings drift")
	}
	allowed := map[int]bool{}
	for _, id := range ids {
		allowed[id] = true
	}
	for id := -1; id <= 127; id++ {
		t.Run(fmt.Sprintf("hero_%d", id), func(t *testing.T) {
			s := NewService(newMemory())
			a, b := creds(t, s), creds(t, s)
			in := roomInput()
			in.Version, in.RosterID, in.Hero = heros28Version, heros28Roster, 1
			room, e := s.CreateRoom(ctx, a.Token, in)
			room = must(t, room, e)
			_, e = s.JoinRoom(ctx, b.Token, JoinRoom{room.Code, heros28Version, 1, heros28Roster})
			if e != nil {
				t.Fatal(e)
			}
			if e = s.Answer(ctx, b.Token, room.ID, heros28Version, Description{"answer", "v=0\r\n"}); e != nil {
				t.Fatal(e)
			}
			_, e = s.BeginSelection(ctx, a.Token, room.ID, BeginSelection{heros28Version, "begin", firstEpoch, "", ""})
			if e != nil {
				t.Fatal(e)
			}
			check := func(e error) {
				t.Helper()
				if allowed[id] {
					if e != nil {
						t.Fatal(e)
					}
				} else {
					expectStatus(t, e, 400)
				}
			}
			for _, c := range []Credentials{a, b} {
				_, e = s.LockSelection(ctx, c.Token, room.ID, LockSelection{heros28Version, "lock", firstEpoch, id})
				check(e)
			}
			for seat, pair := range [][2]int{{id, 1}, {1, id}} {
				key := fmt.Sprintf("heros28_boundary_%02d", seat)
				_, e = s.CreatePVE(ctx, a.Token, PVERequest{key, heros28Version, pair[0], pair[1], "normal", heros28Roster})
				check(e)
				for _, transport := range []string{"local", "broadcastchannel"} {
					_, e = s.CreateLocal(ctx, a.Token, LocalRequest{key + transport, heros28Version, pair[0], pair[1], transport, heros28Roster})
					check(e)
				}
			}
		})
	}
}

// Real handler contract with a memory store: synthetic reports, not browser/P2P acceptance.
func TestHeros28HTTPSelectionLifecycle(t *testing.T) {
	testHeros28HTTPSelectionLifecycle(t, heros28Version)
}

func TestHeros28RetainedSelectionLifecycle(t *testing.T) {
	testHeros28HTTPSelectionLifecycle(t, heros28RetainedVersion)
}

func testHeros28HTTPSelectionLifecycle(t *testing.T, version string) {
	for _, pair := range [][2]int{{0, 2}, {6, 10}, {14, 19}} {
		t.Run(fmt.Sprint(pair), func(t *testing.T) {
			s := NewService(newMemory())
			h := &Handler{Service: s}
			a, b, outsider := creds(t, s), creds(t, s), creds(t, s)
			callNumber := 0
			call := func(method, path, token string, body any, status int, out any) {
				t.Helper()
				raw, _ := json.Marshal(body)
				req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(raw))
				callNumber++
				req.RemoteAddr = fmt.Sprintf("192.0.2.%d:12345", callNumber)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)
				w := httptest.NewRecorder()
				h.ServeHTTP(w, req)
				if w.Code != status {
					t.Fatalf("%s %s status=%d want=%d body=%s", method, path, w.Code, status, w.Body)
				}
				if out != nil {
					if e := json.Unmarshal(w.Body.Bytes(), out); e != nil {
						t.Fatal(e)
					}
				}
			}
			in := roomInput()
			in.Version, in.RosterID, in.Hero = version, heros28Roster, 1
			call("POST", "/rooms", "", in, 401, nil)
			for _, rid := range []string{"", LegacyRosterID, "arena-heros22-v1", "arena-first22-46-v1"} {
				bad := in
				bad.RosterID = rid
				call("POST", "/rooms", a.Token, bad, 409, nil)
			}
			bad := in
			bad.RosterID = RegistryVersion
			call("POST", "/rooms", a.Token, bad, 400, nil)
			for _, version := range []string{"duel-unapproved-build", "duel-64714553cccdaf8cc362", "duel-e63dafb5ae2070a90f8b"} {
				bad := in
				bad.Version = version
				call("POST", "/rooms", a.Token, bad, 409, nil)
			}
			var room RoomView
			call("POST", "/rooms", a.Token, in, 201, &room)
			path := "/rooms/" + room.ID
			call("POST", "/rooms/join", b.Token, JoinRoom{room.Code, version, 1, "arena-heros22-v1"}, 409, nil)
			call("POST", "/rooms/join", b.Token, JoinRoom{room.Code, version, 1, heros28Roster}, 200, nil)
			call("POST", path+"/answer", b.Token, map[string]any{"version": version, "answer": Description{"answer", "v=0\r\n"}}, 200, nil)
			begin := BeginSelection{version, "begin", firstEpoch, "", ""}
			call("POST", path+"/selection", outsider.Token, begin, 403, nil)
			call("POST", path+"/selection", b.Token, begin, 403, nil)
			call("POST", path+"/selection", a.Token, begin, 200, nil)
			call("POST", path+"/selection", a.Token, LockSelection{version, "lock", firstEpoch, 11}, 400, nil)
			call("POST", path+"/matches", a.Token, CreateMatch{"heros28_match_0001", version, firstEpoch}, 409, nil)
			for seat, c := range []Credentials{a, b} {
				lock := LockSelection{version, "lock", firstEpoch, pair[seat]}
				call("POST", path+"/selection", c.Token, lock, 200, nil)
				call("POST", path+"/selection", c.Token, lock, 200, nil)
				lock.Hero = 3
				call("POST", path+"/selection", c.Token, lock, 409, nil)
			}
			call("POST", path+"/selection", a.Token, begin, 200, &room)
			if room.Selection.Locked != [2]bool{true, true} {
				t.Fatal("begin replay erased locks")
			}
			request := CreateMatch{"heros28_match_0001", version, firstEpoch}
			var match, replay Match
			call("POST", path+"/matches", b.Token, request, 403, nil)
			call("POST", path+"/matches", a.Token, request, 201, &match)
			call("POST", path+"/matches", a.Token, request, 201, &replay)
			if replay.ID != match.ID || match.Players[0].Hero != pair[0] || match.Players[1].Hero != pair[1] || match.RosterID != heros28Roster || match.RegistryVersion != RegistryVersion {
				t.Fatal("snapshot/replay drift")
			}
			ready := map[string]string{"version": version}
			mp := "/matches/" + match.ID
			call("POST", mp+"/ready", outsider.Token, ready, 403, nil)
			call("POST", mp+"/ready", a.Token, ready, 200, &replay)
			if replay.Status != "awaiting_ready" {
				t.Fatal("unilateral start")
			}
			call("POST", mp+"/ready", b.Token, ready, 200, &replay)
			if replay.Status != "in_progress" {
				t.Fatal("bilateral start failed")
			}
			next := BeginSelection{version, "begin", nextEpoch, firstEpoch, match.ID}
			call("POST", path+"/selection", a.Token, next, 409, nil)
			report := completed()
			report.Version = version
			call("POST", mp+"/results", a.Token, report, 200, &replay)
			if replay.Status != "pending" {
				t.Fatal("single report promoted")
			}
			call("POST", mp+"/results", b.Token, report, 200, &replay)
			if replay.Status != "confirmed" || replay.Trust != "peer_agreement" {
				t.Fatal("report reconciliation failed")
			}
			snapshot := digest(replay)
			call("POST", path+"/selection", a.Token, next, 200, &room)
			if room.Selection.Locked != [2]bool{} {
				t.Fatal("rematch retained locks")
			}
			for seat, c := range []Credentials{a, b} {
				call("POST", path+"/selection", c.Token, LockSelection{version, "lock", nextEpoch, pair[1-seat]}, 200, nil)
			}
			var second Match
			call("POST", path+"/matches", a.Token, CreateMatch{"heros28_match_0002", version, nextEpoch}, 201, &second)
			if second.ID == match.ID || second.RoomID != match.RoomID || second.Players[0].Hero != pair[1] || second.Players[1].Hero != pair[0] {
				t.Fatal("rematch identity drift")
			}
			call("POST", path+"/selection", a.Token, begin, 409, nil)
			call("GET", mp, a.Token, nil, 200, &replay)
			if digest(replay) != snapshot {
				t.Fatal("historical result changed")
			}
		})
	}
}
