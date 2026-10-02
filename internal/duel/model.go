package duel

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"
)

const RoomTTL = 10 * time.Minute
const SessionTTL = 24 * time.Hour
const MatchTTL = 20 * time.Minute
const ReportTTL = 2 * time.Minute

var ErrMissing = errors.New("not found")
var ErrExists = errors.New("already exists")

type Fault struct {
	Status  int
	Message string
}

func (e *Fault) Error() string { return e.Message }
func bad(s string) error       { return &Fault{400, s} }
func conflict(s string) error  { return &Fault{409, s} }
func forbidden() error         { return &Fault{403, "not a participant"} }
func millis(t time.Time) int64 { return t.UTC().UnixMilli() }
func randomID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func hash(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func randomCode() (string, error) {
	n, e := rand.Int(rand.Reader, big.NewInt(1000000))
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

var codePattern = regexp.MustCompile(`^[0-9]{6}$`)
var keyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{16,80}$`)
var versionPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,100}$`)

type Session struct {
	PlayerID string `json:"playerId"`
	Expires  int64  `json:"expires"`
}
type Credentials struct {
	PlayerID string `json:"playerId"`
	Token    string `json:"token"`
	Expires  int64  `json:"expires"`
}
type Description struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

func (d Description) valid(t string) bool {
	return d.Type == t && len(d.SDP) >= 3 && len(d.SDP) <= 40000 && d.SDP[:3] == "v=0"
}

type Policy struct {
	Direction  string  `json:"direction"`
	RTTMS      float64 `json:"rttMs"`
	JitterMS   int     `json:"jitterMs"`
	LossPct    int     `json:"lossPct"`
	MinSamples int     `json:"minSamples"`
	Window     int     `json:"window"`
	MaxAgeMS   int     `json:"maxAgeMs"`
}

func (p Policy) valid() bool {
	return (p.Direction == "above" || p.Direction == "below") && p.RTTMS >= 1 && p.RTTMS <= 2000 && p.JitterMS == 30 && p.LossPct == 5 && p.MinSamples == 24 && p.Window == 30 && p.MaxAgeMS == 3000
}

type Player struct {
	ID   string `json:"id"`
	Hero int    `json:"hero"`
}
type Room struct {
	ID           string       `json:"id"`
	Code         string       `json:"code"`
	Version      string       `json:"version"`
	Policy       Policy       `json:"policy"`
	Players      [2]Player    `json:"players"`
	Tokens       [2]string    `json:"tokens"`
	Offer        *Description `json:"offer"`
	Answer       *Description `json:"answer"`
	Answered     bool         `json:"answered"`
	CreatedAt    int64        `json:"createdAt"`
	Expires      int64        `json:"expires"`
	Closed       bool         `json:"closed"`
	CurrentMatch string       `json:"currentMatch"`
}
type RoomView struct {
	ID           string       `json:"id"`
	Code         string       `json:"code"`
	Version      string       `json:"version"`
	Policy       Policy       `json:"policy"`
	Players      [2]Player    `json:"players"`
	Offer        *Description `json:"offer,omitempty"`
	Answer       *Description `json:"answer,omitempty"`
	Expires      int64        `json:"expires"`
	CurrentMatch string       `json:"currentMatch"`
	Closed       bool         `json:"closed"`
}

func (r Room) view() RoomView {
	return RoomView{r.ID, r.Code, r.Version, r.Policy, r.Players, r.Offer, r.Answer, r.Expires, r.CurrentMatch, r.Closed}
}
func (r Room) seat(token string) int {
	for i, t := range r.Tokens {
		if t != "" && t == token {
			return i
		}
	}
	return -1
}

type CodeRef struct {
	RoomID  string `json:"roomId"`
	Expires int64  `json:"expires"`
}
type CreateRoom struct {
	RequestID string      `json:"requestId"`
	Version   string      `json:"version"`
	Hero      int         `json:"hero"`
	Offer     Description `json:"offer"`
	Policy    Policy      `json:"policy"`
}
type JoinRoom struct {
	Code    string `json:"code"`
	Version string `json:"version"`
	Hero    int    `json:"hero"`
}
type Round struct {
	Number      int `json:"number"`
	Winner      int `json:"winner"`
	RemainingMS int `json:"remainingMs"`
}
type Result struct {
	Version string  `json:"version"`
	Outcome string  `json:"outcome"`
	Rounds  []Round `json:"rounds"`
	Score   [2]int  `json:"score"`
	Winner  int     `json:"winner"`
	Reason  string  `json:"reason"`
}
type Submission struct {
	Digest     string `json:"digest"`
	Result     Result `json:"result"`
	ReceivedAt int64  `json:"receivedAt"`
}
type Match struct {
	ScoreAgreement   string         `json:"scoreAgreement,omitempty"`
	ID               string         `json:"id"`
	RoomID           string         `json:"roomId"`
	RequestID        string         `json:"requestId"`
	Version          string         `json:"version"`
	Mode             string         `json:"mode"`
	AIDifficulty     string         `json:"aiDifficulty,omitempty"`
	Trust            string         `json:"trust"`
	Transport        string         `json:"transport"`
	ParticipantKinds [2]string      `json:"participantKinds"`
	ReporterPlayerID string         `json:"reporterPlayerId,omitempty"`
	Players          [2]Player      `json:"players"`
	Ready            [2]bool        `json:"ready"`
	Status           string         `json:"status"`
	CreatedAt        int64          `json:"createdAt"`
	StartedAt        int64          `json:"startedAt"`
	EndedAt          int64          `json:"endedAt"`
	Deadline         int64          `json:"deadline"`
	Submissions      [2]*Submission `json:"submissions"`
	Score            [2]int         `json:"score"`
	Winner           int            `json:"winner"`
	Reason           string         `json:"reason"`
}

func (m *Match) terminal() bool {
	return m.Status == "confirmed" || m.Status == "disputed" || m.Status == "aborted" || m.Status == "recorded"
}
func (m *Match) expire(now int64) bool {
	if !m.terminal() && now >= m.Deadline {
		m.Status = "aborted"
		m.ScoreAgreement = "unresolved"
		m.Reason = "result_timeout"
		if m.StartedAt == 0 {
			m.Reason = "start_timeout"
		}
		m.EndedAt = now
		return true
	}
	return false
}
func (m Match) validate(r Result) error {
	if r.Version != m.Version {
		return conflict("version mismatch")
	}
	if len(r.Rounds) > 64 {
		return bad("too many rounds")
	}
	score := [2]int{}
	for i, v := range r.Rounds {
		if v.Number != i+1 || v.Winner < -1 || v.Winner > 1 || v.RemainingMS < 0 || v.RemainingMS > 99000 || score[0] >= 2 || score[1] >= 2 {
			return bad("invalid round sequence")
		}
		if v.Winner >= 0 {
			score[v.Winner]++
		}
	}
	if score != r.Score {
		return bad("score does not match rounds")
	}
	if r.Outcome == "completed" {
		if r.Reason != "" || r.Winner < 0 || r.Winner > 1 || score[r.Winner] != 2 || score[1-r.Winner] >= 2 {
			return bad("invalid winner")
		}
	} else if r.Outcome == "aborted" {
		if r.Winner != -1 || score[0] >= 2 || score[1] >= 2 {
			return bad("aborted match cannot award a win")
		}
		switch r.Reason {
		case "left", "disconnect", "cancelled", "version_mismatch":
		default:
			return bad("invalid abort reason")
		}
	} else {
		return bad("invalid outcome")
	}
	return nil
}
func digest(v any) string { b, _ := json.Marshal(v); return hash(string(b)) }
