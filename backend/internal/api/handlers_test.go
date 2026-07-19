package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/saige/backend/internal/config"
	"github.com/saige/backend/internal/store"
)

func testConfig() *config.Config {
	return &config.Config{
		AllowedOrigins:     []string{"http://localhost:5173"},
		IPHashSalt:         "test-salt",
		LowSampleThreshold: 3,
		RateLimitPerHour:   5,
	}
}

// newTestServer builds the handler against a real DB, skipping without one.
func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run handler integration tests")
	}
	ctx := context.Background()
	st, err := store.New(ctx, url)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(st.Close)
	if err := st.Migrate(ctx, "../../../migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewServer(testConfig(), st)
}

func TestHealthNoDB(t *testing.T) {
	// The security-headers and CORS middleware must apply even without a DB
	// call; exercise them via an OPTIONS preflight, which never touches the DB.
	srv := NewServer(testConfig(), nil)
	req := httptest.NewRequest(http.MethodOptions, "/api/submissions", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("CORS origin = %q", got)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("missing nosniff header, got %q", got)
	}
}

func TestSubmitHappyPath(t *testing.T) {
	srv := newTestServer(t)
	body := `{"email":"happy@example.com","anonymous":true,"ratings":{"EA":2,"IN":1}}`
	rec := doJSON(srv, http.MethodPost, "/api/submissions", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s, want 201", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "created" {
		t.Errorf("status field = %v", resp["status"])
	}
}

func TestSubmitValidationErrors(t *testing.T) {
	srv := newTestServer(t)
	// Missing email + non-anonymous without name.
	body := `{"email":"","anonymous":false,"ratings":{}}`
	rec := doJSON(srv, http.MethodPost, "/api/submissions", body)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body=%s, want 422", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "email") {
		t.Errorf("expected email field error, got %s", rec.Body.String())
	}
}

func TestSubmitHoneypotSilentlyAccepted(t *testing.T) {
	srv := newTestServer(t)
	body := `{"email":"bot@example.com","anonymous":true,"website":"http://spam","ratings":{}}`
	rec := doJSON(srv, http.MethodPost, "/api/submissions", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("honeypot status = %d, want 201 (silent accept)", rec.Code)
	}
	// Verify nothing was persisted for the bot address.
	areas := doJSON(srv, http.MethodGet, "/api/research-areas", "")
	if strings.Contains(areas.Body.String(), "bot@example.com") {
		t.Error("email must never appear in API output")
	}
}

func TestResearchAreasNeverLeakEmail(t *testing.T) {
	srv := newTestServer(t)
	_ = doJSON(srv, http.MethodPost, "/api/submissions",
		`{"email":"secret@example.com","anonymous":false,"name":"Grace","ratings":{"EA":3}}`)
	rec := doJSON(srv, http.MethodGet, "/api/research-areas", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "secret@example.com") {
		t.Error("email leaked in /api/research-areas response")
	}
}

func doJSON(srv http.Handler, method, path, body string) *httptest.ResponseRecorder {
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	r.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, r)
	return rec
}
