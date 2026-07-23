package api

import (
	"strconv"
	"strings"
	"testing"
)

func intp(i int) *int { return &i }

const validToken = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"

func baseReq() FeedbackRequest {
	return FeedbackRequest{
		AgendaID:       "EA1",
		Familiarity:    intp(2),
		AreaMaturity:   "Untested",
		SubmitterToken: validToken,
		IsAnonymous:    true,
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*FeedbackRequest)
		wantField string // "" means the submission must be accepted
	}{
		{"minimal anonymous submission", func(*FeedbackRequest) {}, ""},
		{"missing all targets", func(r *FeedbackRequest) { r.AgendaID = "" }, "agenda_id"},
		{"unknown agenda id", func(r *FeedbackRequest) { r.AgendaID = "ZZ9" }, "agenda_id"},
		{"agenda id is trimmed", func(r *FeedbackRequest) { r.AgendaID = "  SO3  " }, ""},
		{"cell target", func(r *FeedbackRequest) {
			r.AgendaID = ""
			r.AreaTag = "IN"
			r.ProblemID = "P4"
		}, ""},
		{"cell with unknown area", func(r *FeedbackRequest) {
			r.AgendaID = ""
			r.AreaTag = "ZZ"
			r.ProblemID = "P4"
		}, "area_tag"},
		{"cell with unknown problem", func(r *FeedbackRequest) {
			r.AgendaID = ""
			r.AreaTag = "IN"
			r.ProblemID = "P99"
		}, "problem_id"},
		{"cell missing the problem", func(r *FeedbackRequest) {
			r.AgendaID = ""
			r.AreaTag = "IN"
		}, "problem_id"},
		{"missing familiarity", func(r *FeedbackRequest) { r.Familiarity = nil }, "familiarity"},
		{"familiarity below range", func(r *FeedbackRequest) { r.Familiarity = intp(-1) }, "familiarity"},
		{"familiarity above range", func(r *FeedbackRequest) { r.Familiarity = intp(5) }, "familiarity"},
		{"familiarity at lower bound", func(r *FeedbackRequest) { r.Familiarity = intp(0) }, ""},
		{"familiarity at expert bound", func(r *FeedbackRequest) { r.Familiarity = intp(4) }, ""},
		{"missing the maturity vote", func(r *FeedbackRequest) { r.AreaMaturity = "" }, "area_maturity"},
		{"maturity vote is trimmed", func(r *FeedbackRequest) { r.AreaMaturity = "  Contested  " }, ""},
		{"invented tier", func(r *FeedbackRequest) { r.AreaMaturity = "Very robust" }, "area_maturity"},
		{"off-axis tier is a legitimate vote", func(r *FeedbackRequest) { r.AreaMaturity = "Never demonstrated" }, ""},
		{"blank is not a tier anyone can vote for", func(r *FeedbackRequest) { r.AreaMaturity = "(blank)" }, "area_maturity"},
		{"subarea vote in the same area", func(r *FeedbackRequest) {
			r.SubareaMaturity = map[string]string{"EA1": "Contested"}
		}, ""},
		{"subarea vote in another area", func(r *FeedbackRequest) {
			r.SubareaMaturity = map[string]string{"IN6": "Contested"}
		}, "subarea_maturity"},
		{"subarea vote for an unknown agenda", func(r *FeedbackRequest) {
			r.SubareaMaturity = map[string]string{"ZZ9": "Contested"}
		}, "subarea_maturity"},
		{"subarea vote for an invented tier", func(r *FeedbackRequest) {
			r.SubareaMaturity = map[string]string{"EA1": "Very robust"}
		}, "subarea_maturity"},
		{"too many subarea votes", func(r *FeedbackRequest) {
			r.SubareaMaturity = make(map[string]string, MaxSubareaVotes+1)
			for i := 0; i <= MaxSubareaVotes; i++ {
				r.SubareaMaturity[strconv.Itoa(i)] = "Untested"
			}
		}, "subarea_maturity"},
		{"notes at limit", func(r *FeedbackRequest) { r.Notes = strings.Repeat("é", MaxNotesRunes) }, ""},
		{"notes over limit", func(r *FeedbackRequest) { r.Notes = strings.Repeat("é", MaxNotesRunes+1) }, "notes"},
		{"token must be a uuid", func(r *FeedbackRequest) { r.SubmitterToken = "not-a-uuid" }, "submitter_token"},
		{"token must be present", func(r *FeedbackRequest) { r.SubmitterToken = "" }, "submitter_token"},
		{"named submission", func(r *FeedbackRequest) {
			r.IsAnonymous = false
			r.Name = "A Researcher"
			r.Email = "researcher@example.org"
			r.ContactConsent = true
		}, ""},
		{"invalid email", func(r *FeedbackRequest) {
			r.IsAnonymous = false
			r.Email = "not-an-email"
		}, "email"},
		{"contact consent without email", func(r *FeedbackRequest) {
			r.IsAnonymous = false
			r.ContactConsent = true
		}, "email"},
		{"name over limit", func(r *FeedbackRequest) {
			r.IsAnonymous = false
			r.Name = strings.Repeat("x", MaxNameRunes+1)
		}, "name"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := baseReq()
			tc.mutate(&req)
			_, err := Validate(req)

			switch {
			case tc.wantField == "" && err != nil:
				t.Fatalf("expected acceptance, got rejection on %q: %s", err.Field, err.Message)
			case tc.wantField != "" && err == nil:
				t.Fatalf("expected rejection on %q, got acceptance", tc.wantField)
			case tc.wantField != "" && err.Field != tc.wantField:
				t.Fatalf("rejected on %q, want %q", err.Field, tc.wantField)
			}
		})
	}
}

// Anonymity is enforced server-side: a client that ticks "anonymous" but still
// sends a name, an email and contact consent must not have them stored.
func TestValidateStripsIdentityWhenAnonymous(t *testing.T) {
	req := baseReq()
	req.IsAnonymous = true
	req.Name = "A Researcher"
	req.Email = "researcher@example.org"
	req.ContactConsent = true
	req.ReuseConsent = true

	clean, err := Validate(req)
	if err != nil {
		t.Fatalf("unexpected rejection: %v", err)
	}
	if clean.Name != "" || clean.Email != "" {
		t.Errorf("identity survived anonymisation: name=%q email=%q", clean.Name, clean.Email)
	}
	if clean.ContactConsent {
		t.Error("contact consent survived anonymisation")
	}
	if !clean.ReuseConsent {
		t.Error("reuse consent should be kept for anonymous submissions")
	}
}

func TestValidateKeepsCellTarget(t *testing.T) {
	req := baseReq()
	req.AgendaID = ""
	req.AreaTag = "IN"
	req.ProblemID = "P4"

	clean, err := Validate(req)
	if err != nil {
		t.Fatalf("unexpected rejection: %v", err)
	}
	if clean.AgendaID != "" {
		t.Errorf("cell feedback should not carry an agenda id, got %q", clean.AgendaID)
	}
	if clean.AreaTag != "IN" || clean.ProblemID != "P4" {
		t.Errorf("cell target not preserved: area=%q problem=%q", clean.AreaTag, clean.ProblemID)
	}
}

// The vote is about a research area, so the area a submission lands in is derived
// from the agenda rather than taken from the client.
func TestValidateDerivesAreaFromAgenda(t *testing.T) {
	req := baseReq()
	req.AgendaID = "IN6"

	clean, err := Validate(req)
	if err != nil {
		t.Fatalf("unexpected rejection: %v", err)
	}
	if clean.AreaTag != "IN" {
		t.Errorf("area tag = %q, want IN derived from IN6", clean.AreaTag)
	}
	if clean.ProblemID != "" {
		t.Errorf("agenda feedback should carry no problem, got %q", clean.ProblemID)
	}

	// A client that disagrees with the derivation is reporting a bug, not an opinion.
	req.AreaTag = "EA"
	if _, err := Validate(req); err == nil || err.Field != "area_tag" {
		t.Fatalf("expected rejection on area_tag, got %v", err)
	}
}

// An agenda a reader left unrated is not a vote of "no opinion", it is no vote.
func TestValidateDropsBlankSubareaVotes(t *testing.T) {
	req := baseReq()
	req.SubareaMaturity = map[string]string{"EA1": "Contested", "EA2": "", "EA3": "   "}

	clean, err := Validate(req)
	if err != nil {
		t.Fatalf("unexpected rejection: %v", err)
	}
	if len(clean.SubareaMaturity) != 1 || clean.SubareaMaturity["EA1"] != "Contested" {
		t.Errorf("subarea votes = %v, want only the answered one", clean.SubareaMaturity)
	}

	req.SubareaMaturity = map[string]string{"EA1": "", "EA2": ""}
	clean, err = Validate(req)
	if err != nil {
		t.Fatalf("unexpected rejection: %v", err)
	}
	if clean.SubareaMaturity != nil {
		t.Errorf("no answered subarea should leave the map nil, got %v", clean.SubareaMaturity)
	}
}

func TestValidateNormalisesToken(t *testing.T) {
	req := baseReq()
	req.SubmitterToken = strings.ToUpper(validToken)
	clean, err := Validate(req)
	if err != nil {
		t.Fatalf("unexpected rejection: %v", err)
	}
	if clean.SubmitterToken != validToken {
		t.Errorf("token = %q, want lowercased %q", clean.SubmitterToken, validToken)
	}
}

func TestIsUUID(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{validToken, true},
		{strings.ToUpper(validToken), true},
		{"", false},
		{"3f2504e04f8911d39a0c0305e82c3301", false},
		{"3f2504e0-4f89-11d3-9a0c-0305e82c330", false},
		{"3f2504e0-4f89-11d3-9a0c-0305e82c33011", false},
		{"3f2504e0_4f89-11d3-9a0c-0305e82c3301", false},
		{"3f2504e0-4f89-11d3-9a0c-0305e82c33zz", false},
	}
	for _, tc := range tests {
		if got := isUUID(tc.in); got != tc.want {
			t.Errorf("isUUID(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestTruncateRunes(t *testing.T) {
	tests := []struct {
		in   string
		n    int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 2, "he"},
		{"héllo", 2, "hé"},
		{"", 3, ""},
		{"héllo", 0, ""},
	}
	for _, tc := range tests {
		if got := truncateRunes(tc.in, tc.n); got != tc.want {
			t.Errorf("truncateRunes(%q, %d) = %q, want %q", tc.in, tc.n, got, tc.want)
		}
	}
}
