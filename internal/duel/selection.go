package duel

import (
	"context"
	"errors"
	"strings"
)

type BeginSelection struct {
	Version         string `json:"version"`
	Action          string `json:"action"`
	Epoch           string `json:"epoch"`
	PreviousEpoch   string `json:"previousEpoch"`
	PreviousMatchID string `json:"previousMatchId"`
}
type LockSelection struct {
	Version string `json:"version"`
	Action  string `json:"action"`
	Epoch   string `json:"epoch"`
	Hero    int    `json:"hero"`
}

func (s *Service) selectionRoom(t Tx, r Room, version string, now int64) error {
	if version != r.Version || !s.registry.selectionEnabled(version) {
		return conflict("selection version mismatch or disabled")
	}
	if r.Closed || !r.Answered || r.Tokens[1] == "" {
		return conflict("room not ready")
	}
	if now >= r.CreatedAt+SessionTTL.Milliseconds() {
		return conflict("room session expired")
	}
	if _, e := s.registry.resolve(r.RosterID, version, 1); e != nil {
		return e
	}
	for seat, tokenHash := range r.Tokens {
		var session Session
		e := t.Get("session", tokenHash, &session)
		if errors.Is(e, ErrMissing) || e == nil && (session.Expires <= now || session.PlayerID != r.Players[seat].ID) {
			return conflict("participant session expired")
		}
		if e != nil {
			return e
		}
	}
	return nil
}

func (s *Service) BeginSelection(ctx context.Context, token, roomID string, in BeginSelection) (out RoomView, err error) {
	if in.Action != "begin" || !epochPattern.MatchString(in.Epoch) || in.PreviousEpoch != "" && !epochPattern.MatchString(in.PreviousEpoch) {
		return out, bad("invalid selection epoch")
	}
	err = s.Store.Run(ctx, func(t Tx) error {
		r, seat, e := s.room(t, token, roomID)
		if e != nil {
			return e
		}
		if seat != 0 {
			return forbidden()
		}
		now := millis(s.Now())
		if e = s.selectionRoom(t, r, in.Version, now); e != nil {
			return e
		}
		// Keep epoch history independently of the moving current match/selection.
		key := reqKey("selection:"+roomID, token, strings.ToLower(in.Epoch))
		var ref requestRef
		if e = t.Get("request", key, &ref); e == nil {
			if ref.Digest != digest(in) || r.Selection == nil || r.Selection.Epoch != in.Epoch {
				return conflict("selection epoch reused or stale")
			}
			out = r.view()
			return nil
		} else if !errors.Is(e, ErrMissing) {
			return e
		}
		if r.Selection == nil {
			if in.PreviousEpoch != "" || in.PreviousMatchID != "" || r.CurrentMatch != "" {
				return conflict("invalid first selection")
			}
		} else {
			if in.PreviousEpoch != r.Selection.Epoch || in.PreviousMatchID == "" || in.PreviousMatchID != r.Selection.MatchID || in.PreviousMatchID != r.CurrentMatch {
				return conflict("previous selection or match mismatch")
			}
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
		r.Selection = &Selection{SelectionView: SelectionView{Epoch: in.Epoch, PreviousEpoch: in.PreviousEpoch, PreviousMatchID: in.PreviousMatchID}}
		for seat := range r.Players {
			r.Players[seat].Hero = 1
		}
		if e = t.Put("room", r.ID, r, r.Expires); e != nil {
			return e
		}
		if e = t.Insert("request", key, requestRef{r.ID, digest(in)}, r.CreatedAt+SessionTTL.Milliseconds()); e != nil {
			return e
		}
		out = r.view()
		return nil
	})
	return
}

func (s *Service) LockSelection(ctx context.Context, token, roomID string, in LockSelection) (out RoomView, err error) {
	if in.Action != "lock" || !epochPattern.MatchString(in.Epoch) {
		return out, bad("invalid selection epoch")
	}
	err = s.Store.Run(ctx, func(t Tx) error {
		r, seat, e := s.room(t, token, roomID)
		if e != nil {
			return e
		}
		if e = s.selectionRoom(t, r, in.Version, millis(s.Now())); e != nil {
			return e
		}
		if r.Selection == nil || r.Selection.Epoch != in.Epoch {
			return conflict("selection epoch mismatch")
		}
		if _, e = s.registry.resolve(r.RosterID, in.Version, in.Hero); e != nil {
			return e
		}
		if r.Selection.Locked[seat] {
			if r.Players[seat].Hero != in.Hero {
				return conflict("hero already locked")
			}
			// Exact lock replay remains safe even if the host has since allocated a match.
			out = r.view()
			return nil
		}
		if r.Selection.MatchID != "" {
			return conflict("selection already consumed")
		}
		r.Players[seat].Hero = in.Hero
		r.Selection.Locked[seat] = true
		if e = t.Put("room", r.ID, r, r.Expires); e != nil {
			return e
		}
		out = r.view()
		return nil
	})
	return
}
