package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"ai-safety-atlas/backend/internal/config"
	"ai-safety-atlas/backend/internal/store"
)

const testOrigin = "https://atlas.pages.dev"

type fakeStore struct {
	mu        sync.Mutex
	inserted  []store.Feedback
	insertErr error
}

func (f *fakeStore) InsertFeedback(_ context.Context, fb store.Feedback) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.insertErr != nil {
		return f.insertErr
	}
	f.inserted = append(f.inserted, fb)
	return nil
}

func (f *fakeStore) Summary(_ context.Context, iteration int) (store.Summary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return store.Summary{
		Iteration: iteration,
		Totals:    store.Totals{TotalSubmissions: len(f.inserted), UniqueSubmitters: len(f.inserted)},
		PerArea:   map[string]*store.RatingSummary{},
		PerAgenda: map[string]*store.RatingSummary{},
	}, nil
}

func (f *fakeStore) Export(context.Context, int) ([]store.ExportRow, error) {
	id := "EA1"
	return []store.ExportRow{{ID: 1, AgendaID: &id}}, nil
}

func (f *fakeStore) Ping(context.Context) error { return nil }

func (f *fakeStore) rows() []store.Feedback {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]store.Feedback(nil), f.inserted...)
}

func testServer(t *testing.T, st Storage) http.Handler {
	t.Helper()
	cfg := config.Config{
		AllowedOrigin: testOrigin,
		AdminToken:    "s3cret-token",
		Port:          "8080",
		IPHashSalt:    "test-salt",
	}
	return New(cfg, st).Handler()
}

func post(t *testing.T, h http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "203.0.113.9")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestPostFeedbackStoresSubmission(t *testing.T) {
	fs := &fakeStore{}
	h := testServer(t, fs)

	rec := post(t, h, `{"agenda_id":"IN6","familiarity":3,
		"area_maturity":"Early / partial","subarea_maturity":{"IN6":"Untested"},
		"notes":"the API-only team failing is the interesting part",
		"is_anonymous":true,"contact_consent":false,"reuse_consent":true,
		"submitter_token":"`+validToken+`","website":""}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rows := fs.rows()
	if len(rows) != 1 {
		t.Fatalf("stored %d rows, want 1", len(rows))
	}
	got := rows[0]
	if got.AgendaID != "IN6" || got.Familiarity != 3 {
		t.Errorf("stored %+v", got)
	}
	// The area is derived from the agenda, not asked for: the vote is about the
	// area, so which area it lands in is not the client's choice.
	if got.AreaTag != "IN" {
		t.Errorf("area tag = %q, want IN derived from the agenda id", got.AreaTag)
	}
	if got.AreaMaturity != "Early / partial" || got.SubareaMaturity["IN6"] != "Untested" {
		t.Errorf("votes not stored: area=%q subareas=%v", got.AreaMaturity, got.SubareaMaturity)
	}
	if got.IPHash == "" || strings.Contains(got.IPHash, "203.0.113.9") {
		t.Errorf("ip must be stored hashed, got %q", got.IPHash)
	}
}

func TestPostFeedbackStoresCellSubmission(t *testing.T) {
	fs := &fakeStore{}
	h := testServer(t, fs)

	rec := post(t, h, `{"area_tag":"IN","problem_id":"P4","familiarity":2,
		"area_maturity":"Untested","subarea_maturity":{"IN6":"Contested","IN2":""},
		"notes":"you are missing an agenda here","is_anonymous":true,
		"submitter_token":"`+validToken+`","website":""}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rows := fs.rows()
	if len(rows) != 1 {
		t.Fatalf("stored %d rows, want 1", len(rows))
	}
	got := rows[0]
	if got.AgendaID != "" || got.AreaTag != "IN" || got.ProblemID != "P4" {
		t.Errorf("cell target not stored: %+v", got)
	}
	// An agenda the reader left blank is not a vote and must not be stored as one.
	if len(got.SubareaMaturity) != 1 || got.SubareaMaturity["IN6"] != "Contested" {
		t.Errorf("subarea votes = %v, want only the one that was answered", got.SubareaMaturity)
	}
}

func TestPostFeedbackHoneypotIsSilentlyAccepted(t *testing.T) {
	fs := &fakeStore{}
	h := testServer(t, fs)

	rec := post(t, h, `{"agenda_id":"EA1","familiarity":1,"area_maturity":"Untested","is_anonymous":true,
		"submitter_token":"`+validToken+`","website":"http://spam.example"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("honeypot should look like success, got %d", rec.Code)
	}
	if n := len(fs.rows()); n != 0 {
		t.Fatalf("honeypot submission was stored (%d rows)", n)
	}
}

func TestPostFeedbackRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"unknown agenda", `{"agenda_id":"XX1","familiarity":1,"area_maturity":"Untested","submitter_token":"` + validToken + `"}`, http.StatusBadRequest},
		{"familiarity out of range", `{"agenda_id":"EA1","familiarity":9,"area_maturity":"Untested","submitter_token":"` + validToken + `"}`, http.StatusBadRequest},
		{"missing the maturity vote", `{"agenda_id":"EA1","familiarity":1,"submitter_token":"` + validToken + `"}`, http.StatusBadRequest},
		{"invented tier", `{"agenda_id":"EA1","familiarity":1,"area_maturity":"Excellent","submitter_token":"` + validToken + `"}`, http.StatusBadRequest},
		{"subarea vote outside the area", `{"agenda_id":"EA1","familiarity":1,"area_maturity":"Untested","subarea_maturity":{"IN6":"Robust (small scale)"},"submitter_token":"` + validToken + `"}`, http.StatusBadRequest},
		{"retired agree/disagree field", `{"agenda_id":"EA1","familiarity":1,"area_maturity":"Untested","agree_with_rating":"agree","submitter_token":"` + validToken + `"}`, http.StatusBadRequest},
		{"malformed json", `{"agenda_id":`, http.StatusBadRequest},
		{"unknown field", `{"agenda_id":"EA1","familiarity":1,"submitter_token":"` + validToken + `","admin":true}`, http.StatusBadRequest},
		{"oversized body", `{"agenda_id":"EA1","familiarity":1,"submitter_token":"` + validToken + `","notes":"` + strings.Repeat("x", MaxBodyBytes) + `"}`, http.StatusRequestEntityTooLarge},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := &fakeStore{}
			rec := post(t, testServer(t, fs), tc.body)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tc.want, rec.Body.String())
			}
			if n := len(fs.rows()); n != 0 {
				t.Fatalf("invalid submission was stored (%d rows)", n)
			}
		})
	}
}

func TestPostFeedbackRateLimited(t *testing.T) {
	h := testServer(t, &fakeStore{})
	body := `{"agenda_id":"EA1","familiarity":1,"area_maturity":"Untested","is_anonymous":true,"submitter_token":"` + validToken + `"}`

	for i := range 5 {
		if rec := post(t, h, body); rec.Code != http.StatusCreated {
			t.Fatalf("request %d: status = %d", i+1, rec.Code)
		}
	}
	rec := post(t, h, body)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("429 should carry Retry-After")
	}
}

func TestSummaryIsPublicAndAggregateOnly(t *testing.T) {
	fs := &fakeStore{}
	h := testServer(t, fs)

	req := httptest.NewRequest(http.MethodGet, "/api/summary", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding summary: %v", err)
	}
	for _, forbidden := range []string{"submitter_email", "submitter_name", "notes", "ip_hash"} {
		if strings.Contains(rec.Body.String(), forbidden) {
			t.Errorf("summary leaked %q", forbidden)
		}
	}
	for _, key := range []string{"per_area", "per_agenda"} {
		if _, ok := body[key]; !ok {
			t.Errorf("summary should always carry %s", key)
		}
	}
}

func TestExportRequiresAdminToken(t *testing.T) {
	h := testServer(t, &fakeStore{})

	tests := []struct {
		name   string
		header string
		want   int
	}{
		{"no header", "", http.StatusUnauthorized},
		{"wrong token", "Bearer nope", http.StatusUnauthorized},
		{"not bearer", "s3cret-token", http.StatusUnauthorized},
		{"correct token", "Bearer s3cret-token", http.StatusOK},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/admin/export?format=json", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestCORSOnlyAllowsConfiguredOrigin(t *testing.T) {
	h := testServer(t, &fakeStore{})

	tests := []struct {
		origin string
		want   string
	}{
		{testOrigin, testOrigin},
		{"https://evil.example", ""},
		{"", ""},
	}
	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodOptions, "/api/feedback", nil)
		if tc.origin != "" {
			req.Header.Set("Origin", tc.origin)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.want {
			t.Errorf("origin %q: Allow-Origin = %q, want %q", tc.origin, got, tc.want)
		}
		if rec.Code != http.StatusNoContent {
			t.Errorf("preflight status = %d, want 204", rec.Code)
		}
	}
}

func TestHealthAndDegradedMode(t *testing.T) {
	for _, tc := range []struct {
		name  string
		store Storage
		want  string
	}{
		{"with database", &fakeStore{}, "ok"},
		{"without database", nil, "degraded"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := testServer(t, tc.store)
			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d", rec.Code)
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decoding health: %v", err)
			}
			if body["status"] != tc.want {
				t.Fatalf("status = %v, want %v", body["status"], tc.want)
			}
		})
	}
}

func TestFeedbackUnavailableWithoutDatabase(t *testing.T) {
	h := testServer(t, nil)
	rec := post(t, h, `{"agenda_id":"EA1","familiarity":1,"area_maturity":"Untested","is_anonymous":true,"submitter_token":"`+validToken+`"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestHashIPIsSaltedAndStable(t *testing.T) {
	a := hashIP("salt-a", "203.0.113.9")
	b := hashIP("salt-b", "203.0.113.9")
	c := hashIP("salt-a", "203.0.113.9")

	if a == b {
		t.Error("different salts must produce different hashes")
	}
	if a != c {
		t.Error("the same salt and ip must produce the same hash")
	}
	if hashIP("salt-a", "") != "" {
		t.Error("an empty ip should hash to an empty string, not to a hash of the salt")
	}
	if strings.Contains(a, "203.0.113") {
		t.Error("the raw ip must not survive hashing")
	}
}

func TestClientIPPrefersForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.RemoteAddr = "10.0.0.1:5555"
	if got := clientIP(req); got != "10.0.0.1" {
		t.Errorf("clientIP = %q, want 10.0.0.1", got)
	}

	req.Header.Set("X-Forwarded-For", "198.51.100.7, 10.0.0.1")
	if got := clientIP(req); got != "198.51.100.7" {
		t.Errorf("clientIP = %q, want 198.51.100.7", got)
	}
}
