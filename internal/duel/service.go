package duel

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Service struct {
	Store Store
	Now   func() time.Time
	Code  func() (string, error)
}

func NewService(store Store) *Service { return &Service{store, time.Now, randomCode} }
func (s *Service) NewSession(ctx context.Context) (c Credentials, err error) {
	now := millis(s.Now())
	c = Credentials{randomID(), randomID(), now + SessionTTL.Milliseconds()}
	err = s.Store.Run(ctx, func(t Tx) error { return t.Insert("session", hash(c.Token), Session{c.PlayerID, c.Expires}, c.Expires) })
	return
}
func (s *Service) auth(t Tx, token string) (Session, error) {
	var v Session
	if len(token) != 64 {
		return v, &Fault{401, "invalid session"}
	}
	e := t.Get("session", hash(token), &v)
	if errors.Is(e, ErrMissing) || e == nil && v.Expires <= millis(s.Now()) {
		return v, &Fault{401, "session expired"}
	}
	return v, e
}

type requestRef struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}

func reqKey(scope, token, key string) string { return hash(scope + ":" + token + ":" + key) }
func (s *Service) CreateRoom(ctx context.Context, token string, in CreateRoom) (out RoomView, err error) {
	if !keyPattern.MatchString(in.RequestID) || !versionPattern.MatchString(in.Version) || in.Hero < 0 || in.Hero >= 20 || !in.Offer.valid("offer") || !in.Policy.valid() {
		return out, bad("invalid room")
	}
	err = s.Store.Run(ctx, func(t Tx) error {
		session, e := s.auth(t, token)
		if e != nil {
			return e
		}
		key := reqKey("room", token, in.RequestID)
		var ref requestRef
		if e = t.Get("request", key, &ref); e == nil {
			if ref.Digest != digest(in) {
				return conflict("requestId reused with different body")
			}
			var r Room
			if e = t.Get("room", ref.ID, &r); e != nil {
				return e
			}
			out = r.view()
			return nil
		} else if !errors.Is(e, ErrMissing) {
			return e
		}
		now := millis(s.Now())
		r := Room{ID: randomID(), Version: in.Version, Policy: in.Policy, Offer: &in.Offer, CreatedAt: now, Expires: now + RoomTTL.Milliseconds()}
		r.Players[0] = Player{session.PlayerID, in.Hero}
		r.Tokens[0] = hash(token)
		for attempt := 0; attempt < 32; attempt++ {
			code, e := s.Code()
			if e != nil {
				return e
			}
			var old CodeRef
			e = t.Get("code", code, &old)
			if e == nil && old.Expires <= now {
				if e = t.Delete("code", code); e != nil {
					return e
				}
			} else if e != nil && !errors.Is(e, ErrMissing) {
				return e
			}
			e = t.Insert("code", code, CodeRef{r.ID, r.Expires}, r.Expires)
			if errors.Is(e, ErrExists) {
				continue
			}
			if e != nil {
				return e
			}
			r.Code = code
			break
		}
		if r.Code == "" {
			return &Fault{503, "room code capacity exhausted"}
		}
		if e = t.Insert("room", r.ID, r, r.Expires); e != nil {
			return e
		}
		if e = t.Insert("request", key, requestRef{r.ID, digest(in)}, now+SessionTTL.Milliseconds()); e != nil {
			return e
		}
		out = r.view()
		return nil
	})
	return
}
func (s *Service) JoinRoom(ctx context.Context, token string, in JoinRoom) (out RoomView, err error) {
	if !codePattern.MatchString(in.Code) || in.Hero < 0 || in.Hero >= 20 {
		return out, bad("code must contain exactly six digits")
	}
	err = s.Store.Run(ctx, func(t Tx) error {
		session, e := s.auth(t, token)
		if e != nil {
			return e
		}
		var c CodeRef
		e = t.Get("code", in.Code, &c)
		if errors.Is(e, ErrMissing) || e == nil && c.Expires <= millis(s.Now()) {
			return &Fault{404, "room unavailable"}
		}
		if e != nil {
			return e
		}
		var r Room
		if e = t.Get("room", c.RoomID, &r); e != nil {
			return e
		}
		if r.Closed || r.Expires <= millis(s.Now()) {
			return &Fault{404, "room unavailable"}
		}
		if r.Version != in.Version {
			return conflict("version mismatch")
		}
		if r.Tokens[0] == hash(token) {
			return conflict("cannot join your own room")
		}
		if r.Tokens[1] != "" {
			if r.Tokens[1] != hash(token) {
				return conflict("room full")
			}
			if r.Players[1].Hero != in.Hero {
				return conflict("hero already locked")
			}
			out = r.view()
			return nil
		}
		r.Tokens[1] = hash(token)
		r.Players[1] = Player{session.PlayerID, in.Hero}
		if e = t.Put("room", r.ID, r, r.Expires); e != nil {
			return e
		}
		out = r.view()
		return nil
	})
	return
}
func (s *Service) room(t Tx, token, id string) (r Room, seat int, err error) {
	if _, err = s.auth(t, token); err != nil {
		return
	}
	err = t.Get("room", id, &r)
	if err != nil {
		return
	}
	seat = r.seat(hash(token))
	if seat < 0 {
		err = forbidden()
	}
	return
}
func (s *Service) GetRoom(ctx context.Context, token, id string) (out RoomView, err error) {
	err = s.Store.Run(ctx, func(t Tx) error {
		r, _, e := s.room(t, token, id)
		if e != nil {
			return e
		}
		if r.Expires <= millis(s.Now()) {
			r.Offer = nil
			r.Answer = nil
		}
		out = r.view()
		return nil
	})
	return
}
func (s *Service) Answer(ctx context.Context, token, id, version string, d Description) (err error) {
	if !d.valid("answer") {
		return bad("invalid answer")
	}
	return s.Store.Run(ctx, func(t Tx) error {
		r, seat, e := s.room(t, token, id)
		if e != nil {
			return e
		}
		if seat != 1 {
			return forbidden()
		}
		if r.Closed || r.Expires <= millis(s.Now()) {
			return conflict("room expired")
		}
		if version != r.Version {
			return conflict("version mismatch")
		}
		if r.Answer != nil && *r.Answer != d {
			return conflict("answer already locked")
		}
		r.Answer = &d
		r.Answered = true
		return t.Put("room", id, r, r.Expires)
	})
}
func (s *Service) CloseRoom(ctx context.Context, token, id string) error {
	return s.Store.Run(ctx, func(t Tx) error {
		r, _, e := s.room(t, token, id)
		if e != nil {
			return e
		}
		r.Closed = true
		r.Offer = nil
		r.Answer = nil
		// Invalidate only this room's invitation, never a subsequently recycled code.
		var c CodeRef
		if e = t.Get("code", r.Code, &c); e == nil && c.RoomID == r.ID {
			if e = t.Delete("code", r.Code); e != nil {
				return e
			}
		} else if e != nil && !errors.Is(e, ErrMissing) {
			return e
		}
		return t.Put("room", id, r, r.Expires)
	})
}

type CreateMatch struct {
	RequestID string `json:"requestId"`
	Version   string `json:"version"`
}

func (s *Service) CreateMatch(ctx context.Context, token, roomID string, in CreateMatch) (out Match, err error) {
	if !keyPattern.MatchString(in.RequestID) {
		return out, bad("invalid requestId")
	}
	err = s.Store.Run(ctx, func(t Tx) error {
		r, seat, e := s.room(t, token, roomID)
		if e != nil {
			return e
		}
		if seat != 0 {
			return forbidden()
		}
		if r.Version != in.Version {
			return conflict("version mismatch")
		}
		key := reqKey("match:"+roomID, token, in.RequestID)
		var ref requestRef
		if e = t.Get("request", key, &ref); e == nil {
			return t.Get("match", ref.ID, &out)
		} else if !errors.Is(e, ErrMissing) {
			return e
		}
		now := millis(s.Now())
		if r.Closed || r.Tokens[1] == "" || !r.Answered {
			return conflict("room not ready")
		}
		if now >= r.CreatedAt+SessionTTL.Milliseconds() {
			return conflict("room session expired")
		}
		if r.CurrentMatch != "" {
			var prev Match
			if e = t.Get("match", r.CurrentMatch, &prev); e != nil {
				return e
			}
			if prev.expire(now) {
				if e = t.Put("match", prev.ID, prev, prev.Deadline); e != nil {
					return e
				}
			}
			if !prev.terminal() {
				return conflict("previous match still active or pending")
			}
		}
		out = Match{ID: randomID(), RoomID: roomID, RequestID: in.RequestID, Version: r.Version, Players: r.Players, Mode: "pvp", Transport: "webrtc", ParticipantKinds: [2]string{"anonymous_session", "anonymous_session"}, Trust: "peer_agreement", Status: "awaiting_ready", CreatedAt: now, Deadline: now + ReportTTL.Milliseconds(), Winner: -1}
		if e = t.Insert("match", out.ID, out, out.Deadline); e != nil {
			return e
		}
		r.CurrentMatch = out.ID
		if e = t.Put("room", r.ID, r, r.Expires); e != nil {
			return e
		}
		return t.Insert("request", key, requestRef{out.ID, digest(in)}, now+SessionTTL.Milliseconds())
	})
	return
}
func (s *Service) match(t Tx, token, id string) (m Match, seat int, err error) {
	if _, err = s.auth(t, token); err != nil {
		return
	}
	// Matches and rooms are transactionally locked. Deadlocks with concurrent rematch
	// creation are safely rolled back and retried by SQLStore.
	if err = t.Get("match", id, &m); err != nil {
		return
	}
	var r Room
	if err = t.Get("room", m.RoomID, &r); err != nil {
		return
	}
	seat = r.seat(hash(token))
	if seat < 0 {
		err = forbidden()
	}
	return
}
func (s *Service) Ready(ctx context.Context, token, id, version string) (out Match, err error) {
	err = s.Store.Run(ctx, func(t Tx) error {
		m, seat, e := s.match(t, token, id)
		if e != nil {
			return e
		}
		if version != m.Version {
			return conflict("version mismatch")
		}
		now := millis(s.Now())
		if m.expire(now) {
			out = m
			return t.Put("match", id, m, m.Deadline)
		}
		if m.Status != "awaiting_ready" {
			out = m
			return nil
		}
		m.Ready[seat] = true
		if m.Ready[0] && m.Ready[1] {
			m.Status = "in_progress"
			m.StartedAt = now
			m.Deadline = now + MatchTTL.Milliseconds()
		}
		out = m
		return t.Put("match", id, m, m.Deadline)
	})
	return
}
func (s *Service) GetMatch(ctx context.Context, token, id string) (out Match, err error) {
	err = s.Store.Run(ctx, func(t Tx) error {
		m, _, e := s.match(t, token, id)
		if e != nil {
			return e
		}
		if m.expire(millis(s.Now())) {
			if e = t.Put("match", id, m, m.Deadline); e != nil {
				return e
			}
		}
		out = m
		return nil
	})
	return
}
func (s *Service) Submit(ctx context.Context, token, id string, in Result) (out Match, err error) {
	// Normalize empty round lists for deterministic semantic replay.
	if in.Rounds == nil {
		in.Rounds = []Round{}
	}
	err = s.Store.Run(ctx, func(t Tx) error {
		m, seat, e := s.match(t, token, id)
		if e != nil {
			return e
		}
		if e = m.validate(in); e != nil {
			return e
		}
		now := millis(s.Now())
		d := digest(in)
		if old := m.Submissions[seat]; old != nil {
			if old.Digest != d {
				return conflict("submission already locked")
			}
			if m.expire(now) {
				if e = t.Put("match", id, m, m.Deadline); e != nil {
					return e
				}
			}
			out = m
			return nil
		}
		if m.expire(now) {
			out = m
			return t.Put("match", id, m, m.Deadline)
		}
		if m.terminal() || m.Status == "awaiting_ready" {
			return conflict("match does not accept results")
		}
		m.Submissions[seat] = &Submission{d, in, now}
		if m.Trust == "client_reported" {
			m.Status = "recorded"
			m.ScoreAgreement = "single_report"
			m.EndedAt = now
			m.Score = in.Score
			m.Winner = in.Winner
			m.Reason = in.Reason
			if in.Outcome == "aborted" {
				m.Status = "aborted"
			}
		} else if other := m.Submissions[1-seat]; other != nil {
			m.reconcile(now)
		} else {
			m.Status = "pending"
			m.Deadline = now + ReportTTL.Milliseconds()
		}
		out = m
		return t.Put("match", id, m, m.Deadline)
	})
	return
}

type PVERequest struct {
	RequestID    string `json:"requestId"`
	Version      string `json:"version"`
	Hero         int    `json:"hero"`
	OpponentHero int    `json:"opponentHero"`
	AIDifficulty string `json:"aiDifficulty"`
}

func (s *Service) CreatePVE(ctx context.Context, token string, in PVERequest) (out Match, err error) {
	if !keyPattern.MatchString(in.RequestID) || !versionPattern.MatchString(in.Version) || in.Hero < 0 || in.Hero >= 20 || in.OpponentHero < 0 || in.OpponentHero >= 20 {
		return out, bad("invalid pve match")
	}
	switch in.AIDifficulty {
	case "easy", "normal", "hard":
	default:
		return out, bad("invalid AI difficulty")
	}
	err = s.Store.Run(ctx, func(t Tx) error {
		sess, e := s.auth(t, token)
		if e != nil {
			return e
		}
		key := reqKey("pve", token, in.RequestID)
		var ref requestRef
		if e = t.Get("request", key, &ref); e == nil {
			if ref.Digest != digest(in) {
				return conflict("requestId reused")
			}
			return t.Get("match", ref.ID, &out)
		} else if !errors.Is(e, ErrMissing) {
			return e
		}
		now := millis(s.Now())
		r := Room{ID: randomID(), Version: in.Version, CreatedAt: now, Expires: now + SessionTTL.Milliseconds(), Closed: true, Players: [2]Player{{sess.PlayerID, in.Hero}, {"ai", in.OpponentHero}}, Tokens: [2]string{hash(token), ""}}
		out = Match{ID: randomID(), RoomID: r.ID, RequestID: in.RequestID, Version: in.Version, Players: r.Players, Mode: "pve", Transport: "local", ParticipantKinds: [2]string{"anonymous_session", "ai"}, AIDifficulty: in.AIDifficulty, Trust: "client_reported", Status: "in_progress", CreatedAt: now, StartedAt: now, Deadline: now + MatchTTL.Milliseconds(), Winner: -1}
		r.CurrentMatch = out.ID
		if e = t.Insert("room", r.ID, r, r.Expires); e != nil {
			return e
		}
		if e = t.Insert("match", out.ID, out, out.Deadline); e != nil {
			return e
		}
		return t.Insert("request", key, requestRef{out.ID, digest(in)}, now+SessionTTL.Milliseconds())
	})
	return
}

type counter struct {
	Count int `json:"count"`
}

func (s *Service) Limit(ctx context.Context, ip, scope string, perIP, global int) error {
	now := millis(s.Now())
	bucket := now / 60000
	expires := (bucket + 2) * 60000
	return s.Store.Run(ctx, func(t Tx) error {
		for _, b := range []struct {
			key string
			max int
		}{{fmt.Sprintf("%s:%d:global", scope, bucket), global}, {fmt.Sprintf("%s:%d:%s", scope, bucket, hash(ip)), perIP}} {
			var c counter
			e := t.Insert("rate", b.key, c, expires)
			if e != nil && !errors.Is(e, ErrExists) {
				return e
			}
			e = t.Get("rate", b.key, &c)
			if e != nil && !errors.Is(e, ErrMissing) {
				return e
			}
			if c.Count >= b.max {
				return &Fault{429, "rate limit exceeded"}
			}
			c.Count++
			if e = t.Put("rate", b.key, c, expires); e != nil {
				return e
			}
		}
		return nil
	})
}
