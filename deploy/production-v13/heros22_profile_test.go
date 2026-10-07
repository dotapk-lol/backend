package duel

import (
	"fmt"
	"reflect"
	"testing"
)

// Loaded only through the production profile test overlay; no database/network.
func TestHeros22ProductionBoundary(t *testing.T) {
	const rid = "arena-heros22-v1"
	const version = "duel-9431984810f197b393c5"
	ids := []int{1, 3, 4, 5, 7, 8, 9, 15, 17, 18, 28, 31, 32, 36, 50, 55, 57, 58, 62, 71, 81, 82}
	s := NewService(newMemory())
	if len(s.registry.rosters) != 3 || !reflect.DeepEqual(s.registry.byRoster[rid].HeroIDs, ids) || !reflect.DeepEqual(s.registry.byRoster[rid].GameVersions, []string{"duel-27c78aa4cfc8facc8a23", "duel-6b1d12f75aa4bbac4e12", "duel-851e67d77307f479f1fa", version}) {
		t.Fatal("Production22 profile drift")
	}
	a := creds(t, s)
	for _, badRoster := range []string{"", LegacyRosterID, "arena-first22-46-v1"} {
		in := roomInput()
		in.Version = version
		in.Hero = 1
		in.RosterID = badRoster
		if _, e := s.CreateRoom(ctx, a.Token, in); e == nil {
			t.Fatal("New build impersonated old roster", badRoster)
		}
	}
	for _, badVersion := range []string{"duel-e63dafb5ae2070a90f8b", "duel-7e767a8ed2b995465875", "duel-unapproved-build"} {
		in := roomInput()
		in.Version = badVersion
		in.Hero = 1
		in.RosterID = rid
		if _, e := s.CreateRoom(ctx, a.Token, in); e == nil {
			t.Fatal("Wrong build accepted22", badVersion)
		}
	}
	for _, oldVersion := range []string{"duel-e63dafb5ae2070a90f8b", "duel-7e767a8ed2b995465875"} {
		in := roomInput()
		in.Version = oldVersion
		in.Hero = 99
		in.RosterID = "arena-first22-46-v1"
		in.RequestID = "old_roster_request_" + oldVersion
		if _, e := s.CreateRoom(ctx, a.Token, in); e != nil {
			t.Fatal("Old46 regression", e)
		}
	}
	allowed := map[int]bool{}
	for _, id := range ids {
		allowed[id] = true
	}
	for id := 0; id < 127; id++ {
		t.Run(fmt.Sprintf("hero_%d", id), func(t *testing.T) {
			s := NewService(newMemory())
			a, b := creds(t, s), creds(t, s)
			check := func(e error) {
				t.Helper()
				if (e == nil) != allowed[id] {
					t.Fatalf("Hero%d allowed=%v error=%v", id, allowed[id], e)
				}
			}
			in := roomInput()
			in.Version = version
			in.RosterID = rid
			in.Hero = id
			_, e := s.CreateRoom(ctx, a.Token, in)
			check(e)
			in.Hero = 1
			in.RequestID = "join_anchor_request_01"
			r, e := s.CreateRoom(ctx, a.Token, in)
			if e != nil {
				t.Fatal(e)
			}
			_, e = s.JoinRoom(ctx, b.Token, JoinRoom{r.Code, version, id, rid})
			check(e)
			for seat, pair := range [][2]int{{id, 1}, {1, id}} {
				key := fmt.Sprintf("boundary_request_%02d", seat)
				_, e = s.CreatePVE(ctx, a.Token, PVERequest{key, version, pair[0], pair[1], "normal", rid})
				check(e)
				for _, transport := range []string{"local", "broadcastchannel"} {
					_, e = s.CreateLocal(ctx, a.Token, LocalRequest{key + transport, version, pair[0], pair[1], transport, rid})
					check(e)
				}
			}
		})
	}
}
