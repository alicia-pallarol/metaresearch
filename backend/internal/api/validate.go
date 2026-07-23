package api

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"ai-safety-atlas/backend/internal/atlas"
)

// Limits applied to every submission. Kept here so the frontend and the tests
// have one place to read them from.
const (
	MaxNotesRunes = 2000
	MaxNameRunes  = 120
	MaxEmailRunes = 254
	MaxUARunes    = 256
	// MaxSubareaVotes caps subarea_maturity. The largest research area has eight
	// agendas; the ceiling is generous but finite so the map cannot be voted on
	// wholesale in one request.
	MaxSubareaVotes = 32
	// MaxBodyBytes caps the request body well above the field limits above.
	MaxBodyBytes = 16 << 10 // 16 KiB
)

// FeedbackRequest is the wire format of POST /api/feedback.
//
// Two things travel together and should not be confused:
//
//   - The TARGET says where the form was opened: one agenda (agenda_id), or an
//     Area x Problem cell (area_tag + problem_id). Exactly one shape is accepted.
//     It is provenance, which page produced this answer.
//   - The VOTES are what the reader actually claims: area_maturity, always, for
//     the research area as a whole, and optionally subarea_maturity, a tier for
//     individual agendas within it.
//
// Maturity is a property of the research, not of a problem, how much work exists
// and what it has shown, so a vote is never scoped to a problem, even when it was
// cast from a cell.
type FeedbackRequest struct {
	AgendaID  string `json:"agenda_id"`
	AreaTag   string `json:"area_tag"`
	ProblemID string `json:"problem_id"`
	// AreaMaturity is the required vote: one tier from legend.tier_order.
	AreaMaturity string `json:"area_maturity"`
	// SubareaMaturity is agenda id -> tier, for agendas of the voted-on area.
	SubareaMaturity map[string]string `json:"subarea_maturity"`
	Familiarity     *int              `json:"familiarity"`
	Notes           string            `json:"notes"`
	Name            string            `json:"name"`
	Email           string            `json:"email"`
	IsAnonymous     bool              `json:"is_anonymous"`
	ContactConsent  bool              `json:"contact_consent"`
	ReuseConsent    bool              `json:"reuse_consent"`
	SubmitterToken  string            `json:"submitter_token"`
	TurnstileToken  string            `json:"turnstile_token"`
	// Website is a honeypot: real users never see this field, so a non-empty
	// value means a bot filled the form in.
	Website string `json:"website"`
}

// CleanFeedback is a validated submission, ready to store.
type CleanFeedback struct {
	AgendaID  string
	AreaTag   string
	ProblemID string
	// AreaMaturity is the tier voted for the research area as a whole.
	AreaMaturity string
	// SubareaMaturity is agenda id -> tier; nil when the reader rated no agenda.
	SubareaMaturity map[string]string
	Familiarity     int16
	Notes           string
	Name            string
	Email           string
	IsAnonymous     bool
	ContactConsent  bool
	ReuseConsent    bool
	SubmitterToken  string
}

// ValidationError is a field-level rejection, safe to return to the client.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string { return e.Field + ": " + e.Message }

// Validate checks a submission and normalises it.
//
// Anonymity is enforced here rather than trusted from the client: when
// is_anonymous is set, name, email and contact consent are dropped before the
// value ever reaches the database.
func Validate(req FeedbackRequest) (CleanFeedback, *ValidationError) {
	var out CleanFeedback

	// Resolve the target: an agenda, or an Area x Problem cell. An agenda id, when
	// present, wins and defines agenda-level feedback; otherwise the area+problem
	// pair defines cell-level feedback.
	agendaID := strings.TrimSpace(req.AgendaID)
	areaTag := strings.TrimSpace(req.AreaTag)
	problemID := strings.TrimSpace(req.ProblemID)

	switch {
	case agendaID != "":
		if !atlas.KnownAgendaID(agendaID) {
			return out, &ValidationError{"agenda_id", "unknown agenda"}
		}
		out.AgendaID = agendaID
		// The area is derived, never taken on trust: the vote is about the area,
		// so which area it lands in cannot be a client's choice. A client that
		// sends a different one is reporting a bug, not an opinion.
		derived, _ := atlas.AreaOfAgenda(agendaID)
		if areaTag != "" && areaTag != derived {
			return out, &ValidationError{"area_tag", "does not match the agenda's research area"}
		}
		out.AreaTag = derived
	case areaTag != "" || problemID != "":
		if !atlas.KnownAreaTag(areaTag) {
			return out, &ValidationError{"area_tag", "unknown research area"}
		}
		if !atlas.KnownProblemID(problemID) {
			return out, &ValidationError{"problem_id", "unknown problem"}
		}
		out.AreaTag = areaTag
		out.ProblemID = problemID
	default:
		return out, &ValidationError{"agenda_id", "feedback must target an agenda, or an area and a problem"}
	}

	if req.Familiarity == nil {
		return out, &ValidationError{"familiarity", "required"}
	}
	if *req.Familiarity < 0 || *req.Familiarity > 4 {
		return out, &ValidationError{"familiarity", "must be between 0 and 4"}
	}
	out.Familiarity = int16(*req.Familiarity)

	// The vote itself. Required, and always about the research area as a whole,
	// including when the form was opened from a single agenda or a single cell.
	out.AreaMaturity = strings.TrimSpace(req.AreaMaturity)
	if out.AreaMaturity == "" {
		return out, &ValidationError{"area_maturity", "required"}
	}
	if !atlas.KnownTier(out.AreaMaturity) {
		return out, &ValidationError{"area_maturity", "must be one of the six maturity tiers"}
	}

	if len(req.SubareaMaturity) > MaxSubareaVotes {
		return out, &ValidationError{"subarea_maturity", fmt.Sprintf("at most %d agendas may be rated at once", MaxSubareaVotes)}
	}
	for id, tier := range req.SubareaMaturity {
		id, tier = strings.TrimSpace(id), strings.TrimSpace(tier)
		if tier == "" {
			continue // an agenda the reader left blank; not an error
		}
		if !atlas.KnownAgendaID(id) {
			return out, &ValidationError{"subarea_maturity", "unknown agenda: " + id}
		}
		// An agenda can only be rated under its own area, so one submission cannot
		// reach across the map.
		if area, _ := atlas.AreaOfAgenda(id); area != out.AreaTag {
			return out, &ValidationError{"subarea_maturity", id + " is not in this research area"}
		}
		if !atlas.KnownTier(tier) {
			return out, &ValidationError{"subarea_maturity", "unknown tier for " + id}
		}
		if out.SubareaMaturity == nil {
			out.SubareaMaturity = make(map[string]string, len(req.SubareaMaturity))
		}
		out.SubareaMaturity[id] = tier
	}

	out.Notes = strings.TrimSpace(req.Notes)
	if utf8.RuneCountInString(out.Notes) > MaxNotesRunes {
		return out, &ValidationError{"notes", fmt.Sprintf("must be at most %d characters", MaxNotesRunes)}
	}

	token := strings.TrimSpace(req.SubmitterToken)
	if !isUUID(token) {
		return out, &ValidationError{"submitter_token", "must be a uuid"}
	}
	out.SubmitterToken = strings.ToLower(token)

	out.IsAnonymous = req.IsAnonymous
	out.ReuseConsent = req.ReuseConsent

	if out.IsAnonymous {
		// Nothing identifying is kept for an anonymous submission, and there is
		// nobody to contact, so contact consent is meaningless.
		out.Name, out.Email, out.ContactConsent = "", "", false
		return out, nil
	}

	out.Name = strings.TrimSpace(req.Name)
	if utf8.RuneCountInString(out.Name) > MaxNameRunes {
		return out, &ValidationError{"name", fmt.Sprintf("must be at most %d characters", MaxNameRunes)}
	}

	out.Email = strings.TrimSpace(req.Email)
	if out.Email != "" {
		if utf8.RuneCountInString(out.Email) > MaxEmailRunes {
			return out, &ValidationError{"email", "too long"}
		}
		if _, err := mail.ParseAddress(out.Email); err != nil {
			return out, &ValidationError{"email", "not a valid email address"}
		}
	}
	out.ContactConsent = req.ContactConsent

	if out.ContactConsent && out.Email == "" {
		return out, &ValidationError{"email", "required when you ask to be contacted"}
	}

	return out, nil
}

// isUUID accepts the canonical 8-4-4-4-12 hex form, any version.
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, r := range s {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			isHex := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
			if !isHex {
				return false
			}
		}
	}
	return true
}

// truncateRunes cuts s to at most n runes without splitting one.
func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	count := 0
	for i := range s {
		if count == n {
			return s[:i]
		}
		count++
	}
	return s
}
