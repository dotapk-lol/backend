package duel

import (
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

const primaryOrigin = "https://dotapk.lol"
const workerOrigin = "https://dotapk-frontend.skiyo.workers.dev"

func TestParseAllowedOrigins(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want []string
	}{
		{"", []string{}},
		{primaryOrigin, []string{primaryOrigin}},
		{primaryOrigin + ", " + workerOrigin, []string{primaryOrigin, workerOrigin}},
		{primaryOrigin + "," + primaryOrigin, []string{primaryOrigin}},
		{"http://127.0.0.1:4196", []string{"http://127.0.0.1:4196"}},
		{"http://[::1]:4196", []string{"http://[::1]:4196"}},
	} {
		got, err := ParseAllowedOrigins(tc.raw)
		if err != nil || !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("valid configuration rejected: %v", err)
		}
	}
	for _, raw := range []string{
		"*", "null", "https://*.workers.dev", primaryOrigin + ",", "," + workerOrigin,
		primaryOrigin + "/", primaryOrigin + "/path", primaryOrigin + "?", primaryOrigin + "#",
		primaryOrigin + "?q=1", primaryOrigin + "#fragment", "https://user@dotapk.lol",
		"ftp://dotapk.lol", "//dotapk.lol", "https://", "https://dotapk.lol:",
		"https://dotapk.lol:bad", primaryOrigin + ",null", primaryOrigin + ",https://bad host",
	} {
		if _, err := ParseAllowedOrigins(raw); err == nil {
			t.Errorf("invalid configuration accepted: %q", raw)
		}
	}
}

func TestExactOriginCORS(t *testing.T) {
	for _, origin := range []string{primaryOrigin, workerOrigin} {
		for _, method := range []string{"GET", "OPTIONS"} {
			h := Handler{Service: NewService(newMemory()), Origin: primaryOrigin + "," + workerOrigin}
			r := httptest.NewRequest(method, "/api/v1/registry", nil)
			r.Header.Set("Origin", origin)
			r.Header.Set("Access-Control-Request-Method", "POST")
			r.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			want := 200
			if method == "OPTIONS" {
				want = 204
				if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" ||
					w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, DELETE, OPTIONS" {
					t.Fatal("preflight contract changed")
				}
			}
			if w.Code != want || w.Header().Get("Access-Control-Allow-Origin") != origin ||
				w.Header().Get("Vary") != "Origin" || w.Header().Get("Access-Control-Allow-Credentials") != "" {
				t.Fatalf("CORS mismatch: status %d", w.Code)
			}
		}
	}
	h := Handler{Service: NewService(newMemory()), Origin: primaryOrigin}
	for _, origin := range []string{"", primaryOrigin, workerOrigin} {
		r := httptest.NewRequest("GET", "/api/v1/registry", nil)
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		want := 200
		if origin == workerOrigin {
			want = 403
		}
		if w.Code != want || origin == "" && w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatal("original single/absent Origin behavior changed")
		}
	}
}

func TestRejectedOriginCannotReachStore(t *testing.T) {
	for _, origin := range []string{
		"null", "https://evil.invalid", "https://other.skiyo.workers.dev",
		workerOrigin + ".evil.invalid", "https://sub.dotapk.lol", "http://dotapk.lol",
		primaryOrigin + ":443", primaryOrigin + "/", primaryOrigin + "," + workerOrigin,
	} {
		for _, method := range []string{"GET", "OPTIONS", "POST"} {
			store := newMemory()
			h := Handler{Service: NewService(store), Origin: primaryOrigin + "," + workerOrigin}
			r := httptest.NewRequest(method, "/api/v1/sessions", strings.NewReader("{}"))
			r.Header.Set("Origin", origin)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 403 || w.Header().Get("Access-Control-Allow-Origin") != "" || len(store.rows) != 0 {
				t.Fatal("rejected origin reached storage or received permission")
			}
		}
	}
	for _, configuration := range []string{primaryOrigin + ",null", "*", "null", primaryOrigin + ","} {
		h := Handler{Origin: configuration}
		r := httptest.NewRequest("OPTIONS", "/api/v1/sessions", nil)
		r.Header.Set("Origin", primaryOrigin)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatal("invalid configuration did not fail closed")
		}
	}
	h := Handler{Origin: primaryOrigin + "," + workerOrigin}
	r := httptest.NewRequest("OPTIONS", "/api/v1/sessions", nil)
	r.Header.Add("Origin", primaryOrigin)
	r.Header.Add("Origin", workerOrigin)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("multiple Origin headers accepted")
	}
}

func TestTwoOriginsKeepBearerAndParticipantChecks(t *testing.T) {
	s := NewService(newMemory())
	a, b := creds(t, s), creds(t, s)
	r, err := s.CreateRoom(ctx, a.Token, roomInput())
	r = must(t, r, err)
	h := Handler{Service: s, Origin: primaryOrigin + "," + workerOrigin}
	for _, origin := range []string{primaryOrigin, workerOrigin} {
		for _, tc := range []struct {
			token string
			want  int
		}{{"", 401}, {strings.Repeat("0", 64), 401}, {b.Token, 403}, {a.Token, 200}} {
			req := httptest.NewRequest("GET", "/api/v1/rooms/"+r.ID, nil)
			req.Header.Set("Origin", origin)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != tc.want || w.Header().Get("Access-Control-Allow-Origin") != origin ||
				w.Header().Get("Set-Cookie") != "" {
				t.Fatalf("bearer/participant contract changed: %d", w.Code)
			}
		}
	}
}
