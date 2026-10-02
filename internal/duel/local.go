package duel

import (
	"context"
	"errors"
)

// LocalRequest registers two unverified slots controlled by one reporter.
// It never creates a second authenticated participant or accepts a trust field.
type LocalRequest struct {
	RequestID    string `json:"requestId"`
	Version      string `json:"version"`
	Hero         int    `json:"hero"`
	OpponentHero int    `json:"opponentHero"`
	Transport    string `json:"transport"`
	RosterID     string `json:"rosterId,omitempty" wire:"optional-nonempty"`
}

func (s *Service) CreateLocal(ctx context.Context, token string, in LocalRequest) (out Match, err error) {
	if !keyPattern.MatchString(in.RequestID) || !versionPattern.MatchString(in.Version) || (in.Transport != "local" && in.Transport != "broadcastchannel") {
		return out, bad("invalid local PVP match")
	}
	roster, err := s.registry.resolve(in.RosterID, in.Version, in.Hero, in.OpponentHero)
	if err != nil {
		return out, err
	}

	err = s.Store.Run(ctx, func(t Tx) error {
		sess, e := s.auth(t, token)
		if e != nil {
			return e
		}
		key := reqKey("local", token, in.RequestID)
		var ref requestRef
		if e = t.Get("request", key, &ref); e == nil {
			if ref.Digest != digest(in) {
				return conflict("requestId reused with different body")
			}
			return t.Get("match", ref.ID, &out)
		} else if !errors.Is(e, ErrMissing) {
			return e
		}
		now := millis(s.Now())
		// Slot IDs are unique to this match, not persistent human identities. Only the
		// true reporting session is authorized, through the private room token hash.
		r := Room{ID: randomID(), Version: in.Version, RosterID: roster.ID, RegistryVersion: roster.RegistryVersion, CreatedAt: now, Expires: now + SessionTTL.Milliseconds(), Closed: true, Players: [2]Player{{randomID(), in.Hero}, {randomID(), in.OpponentHero}}, Tokens: [2]string{hash(token), ""}}
		out = Match{ID: randomID(), RoomID: r.ID, RequestID: in.RequestID, Version: in.Version, RosterID: roster.ID, RegistryVersion: roster.RegistryVersion, Players: r.Players, Mode: "pvp", Transport: in.Transport, ParticipantKinds: [2]string{"local_slot", "local_slot"}, ReporterPlayerID: sess.PlayerID, Trust: "client_reported", Status: "in_progress", CreatedAt: now, StartedAt: now, Deadline: now + MatchTTL.Milliseconds(), Winner: -1}
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
