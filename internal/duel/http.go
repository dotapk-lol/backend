package duel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type Handler struct {
	Service      *Service
	Origin       string
	TrustedProxy string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	origin := r.Header.Get("Origin")
	if origin != "" && origin != h.Origin {
		replyError(w, &Fault{403, "origin not allowed"})
		return
	}
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", h.Origin)
		w.Header().Set("Vary", "Origin")
	}
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(204)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	r = r.WithContext(ctx)
	if r.URL.Path == "/healthz" && r.Method == "GET" {
		if e := h.Service.Store.Ping(ctx); e != nil {
			replyError(w, &Fault{503, "database unavailable"})
			return
		}
		reply(w, 200, map[string]any{"ok": true, "service": "dota-duel", "contractVersion": "v1.2-abort-reconciliation", "transport": "webrtc", "capabilities": []string{"pvp_peer_agreement", "pve_client_reported", "local_pvp_client_reported"}, "turn": false})
		return
	}
	ip, _, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		ip = r.RemoteAddr
	}
	if h.TrustedProxy != "" && ip == h.TrustedProxy {
		if forwarded := net.ParseIP(r.Header.Get("X-Real-IP")); forwarded != nil {
			ip = forwarded.String()
		}
	}
	if e = h.Service.Limit(ctx, ip, "requests", 180, 1200); e != nil {
		replyError(w, e)
		return
	}
	if r.Method == "POST" {
		scope := ""
		limit := 0
		total := 0
		switch r.URL.Path {
		case "/api/v1/sessions":
			scope = "sessions"
			limit = 10
			total = 60
		case "/api/v1/rooms/join":
			scope = "join"
			limit = 10
			total = 60
		case "/api/v1/rooms":
			scope = "create"
			limit = 6
			total = 60
		}
		if scope != "" {
			if e = h.Service.Limit(ctx, ip, scope, limit, total); e != nil {
				replyError(w, e)
				return
			}
		}
	}
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if r.Header.Get("Authorization") == token {
		token = ""
	}
	var result any
	status := 200
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if r.URL.Path == "/api/v1/sessions" && r.Method == "POST" {
		result, e = h.Service.NewSession(ctx)
		status = 201
	} else if r.URL.Path == "/api/v1/rooms" && r.Method == "POST" {
		var in CreateRoom
		if e = decode(w, r, &in); e == nil {
			result, e = h.Service.CreateRoom(ctx, token, in)
			status = 201
		}
	} else if r.URL.Path == "/api/v1/rooms/join" && r.Method == "POST" {
		var in JoinRoom
		if e = decode(w, r, &in); e == nil {
			result, e = h.Service.JoinRoom(ctx, token, in)
		}
	} else if r.URL.Path == "/api/v1/matches/pve" && r.Method == "POST" {
		var in PVERequest
		if e = decode(w, r, &in); e == nil {
			result, e = h.Service.CreatePVE(ctx, token, in)
			status = 201
		}
	} else if r.URL.Path == "/api/v1/matches/local" && r.Method == "POST" {
		var in LocalRequest
		if e = decode(w, r, &in); e == nil {
			result, e = h.Service.CreateLocal(ctx, token, in)
			status = 201
		}
	} else if len(parts) >= 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "rooms" && len(parts[3]) == 64 {
		id := parts[3]
		switch {
		case len(parts) == 4 && r.Method == "GET":
			result, e = h.Service.GetRoom(ctx, token, id)
		case len(parts) == 4 && r.Method == "DELETE":
			e = h.Service.CloseRoom(ctx, token, id)
			result = map[string]bool{"ok": true}
		case len(parts) == 5 && parts[4] == "answer" && r.Method == "POST":
			var in struct {
				Version string      `json:"version"`
				Answer  Description `json:"answer"`
			}
			if e = decode(w, r, &in); e == nil {
				e = h.Service.Answer(ctx, token, id, in.Version, in.Answer)
				result = map[string]bool{"ok": true}
			}
		case len(parts) == 5 && parts[4] == "matches" && r.Method == "POST":
			var in CreateMatch
			if e = decode(w, r, &in); e == nil {
				result, e = h.Service.CreateMatch(ctx, token, id, in)
				status = 201
			}
		default:
			e = &Fault{404, "route not found"}
		}
	} else if len(parts) >= 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "matches" && len(parts[3]) == 64 {
		id := parts[3]
		switch {
		case len(parts) == 4 && r.Method == "GET":
			result, e = h.Service.GetMatch(ctx, token, id)
		case len(parts) == 5 && parts[4] == "ready" && r.Method == "POST":
			var in struct {
				Version string `json:"version"`
			}
			if e = decode(w, r, &in); e == nil {
				result, e = h.Service.Ready(ctx, token, id, in.Version)
			}
		case len(parts) == 5 && parts[4] == "results" && r.Method == "POST":
			var in Result
			if e = decode(w, r, &in); e == nil {
				result, e = h.Service.Submit(ctx, token, id, in)
			}
		default:
			e = &Fault{404, "route not found"}
		}
	} else {
		e = &Fault{404, "route not found"}
	}
	if e != nil {
		replyError(w, e)
		return
	}
	if m, ok := result.(Match); ok {
		result = publicMatch(m)
	}
	reply(w, status, result)
}
func publicMatch(m Match) any {
	b, _ := json.Marshal(m)
	var v map[string]any
	_ = json.Unmarshal(b, &v)
	delete(v, "submissions")
	v["reported"] = [2]bool{m.Submissions[0] != nil, m.Submissions[1] != nil}
	return v
}
func reply(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func replyError(w http.ResponseWriter, e error) {
	status := 503
	message := "service unavailable"
	var f *Fault
	if errors.As(e, &f) {
		status = f.Status
		message = f.Message
	} else if errors.Is(e, ErrMissing) {
		status = 404
		message = "not found"
	}
	if status == 429 {
		w.Header().Set("Retry-After", "60")
	}
	reply(w, status, map[string]string{"error": message})
}
func decode(w http.ResponseWriter, r *http.Request, out any) error {
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		return &Fault{415, "application/json required"}
	}
	b, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 45000))
	if e != nil {
		return &Fault{413, "body too large"}
	}
	if e = shape(b, reflect.TypeOf(out).Elem()); e != nil {
		return e
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e = d.Decode(out); e != nil {
		return bad("invalid JSON body")
	}
	return nil
}

// Require every declared input field and exact array lengths; encoding/json alone
// accepts missing fields and silently truncates fixed-size arrays.
func shape(b json.RawMessage, t reflect.Type) error {
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		return bad("null field")
	}
	switch t.Kind() {
	case reflect.Struct:
		var obj map[string]json.RawMessage
		if json.Unmarshal(b, &obj) != nil || obj == nil {
			return bad("object required")
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			key := strings.Split(f.Tag.Get("json"), ",")[0]
			v, ok := obj[key]
			if !ok {
				return bad("missing field: " + key)
			}
			if e := shape(v, f.Type); e != nil {
				return e
			}
		}
	case reflect.Array, reflect.Slice:
		var a []json.RawMessage
		if json.Unmarshal(b, &a) != nil {
			return bad("array required")
		}
		if t.Kind() == reflect.Array && len(a) != t.Len() {
			return bad("invalid array length")
		}
		for _, v := range a {
			if e := shape(v, t.Elem()); e != nil {
				return e
			}
		}
	}
	return nil
}
