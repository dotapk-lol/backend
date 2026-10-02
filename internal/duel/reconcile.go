package duel

// reconcile resolves two immutable, individually validated reports. Aborted
// reports describe interruption from different viewpoints, not competing wins.
// Never rewrite either report or its digest while deriving the match summary.
func (m *Match) reconcile(now int64) {
	a, b := m.Submissions[0], m.Submissions[1]
	m.EndedAt = now
	m.Winner = -1
	m.Score = [2]int{}
	m.ScoreAgreement = "unresolved"
	switch {
	case a.Result.Outcome == "aborted" && b.Result.Outcome == "aborted":
		m.Status = "aborted"
		m.Reason = "interrupted"
		if a.Result.Reason == b.Result.Reason {
			m.Reason = a.Result.Reason
		}
		if a.Result.Score == b.Result.Score {
			m.Score = a.Result.Score
			m.ScoreAgreement = "peer_agreed"
		}
	case a.Result.Outcome != b.Result.Outcome:
		// A completed-result claim contradicted by an abort remains an outcome dispute.
		m.Status = "disputed"
		m.Reason = "outcome_conflict"
	case a.Digest != b.Digest:
		m.Status = "disputed"
		m.Reason = "conflicting_reports"
	default:
		m.Status = "confirmed"
		m.Reason = ""
		m.Score = a.Result.Score
		m.Winner = a.Result.Winner
		m.ScoreAgreement = "peer_agreed"
	}
}
