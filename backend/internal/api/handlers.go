package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/saige/backend/internal/store"
)

// handleHealth reports liveness and database connectivity. The frontend pings
// this to detect (and wait out) a cold start on the free hosting tier.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := withTimeout(r, 5*time.Second)
	defer cancel()
	status := "ok"
	code := http.StatusOK
	if err := s.store.Ping(ctx); err != nil {
		status, code = "degraded", http.StatusServiceUnavailable
	}
	writeJSON(w, code, map[string]string{"status": status})
}

// handleResearchAreas returns the framework plus aggregate familiarity stats.
// It exposes only aggregates — never emails or individual submissions.
func (s *Server) handleResearchAreas(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := withTimeout(r, 15*time.Second)
	defer cancel()

	agg, err := s.store.AreaAggregates(ctx, s.cfg.LowSampleThreshold)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to load research areas.")
		return
	}
	writeJSON(w, http.StatusOK, agg)
}

const maxBodyBytes = 64 * 1024

// handleSubmit validates and stores a familiarity submission. A repeat email
// replaces its prior submission. A filled honeypot is silently accepted.
func (s *Server) handleSubmit(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := withTimeout(r, 15*time.Second)
	defer cancel()

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	var req submissionRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		if err == io.EOF {
			writeError(w, http.StatusBadRequest, "Request body is empty.")
			return
		}
		writeError(w, http.StatusBadRequest, "Malformed request body.")
		return
	}

	// Honeypot: a real user cannot fill a hidden field. Pretend success without
	// persisting anything, so bots get no signal.
	if req.Honeypot != "" {
		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"status": "created", "replaced": false,
		})
		return
	}

	tagToID, err := s.store.TagToID(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to load research areas.")
		return
	}

	email, name, ratings, verrs := req.validate(tagToID)
	if len(verrs) > 0 {
		writeError(w, http.StatusUnprocessableEntity, "Please correct the highlighted fields.", verrs...)
		return
	}

	// Rate limit: in-memory window first (cheap), then a durable DB count.
	ipHash := s.hashIP(r)
	if !s.limiter.allow(ipHash) {
		writeError(w, http.StatusTooManyRequests, "Too many submissions from your network. Please try again later.")
		return
	}
	since := time.Now().Add(-time.Hour)
	if n, err := s.store.CountRecentByIPHash(ctx, ipHash, since); err == nil && n >= s.cfg.RateLimitPerHour {
		writeError(w, http.StatusTooManyRequests, "Too many submissions from your network. Please try again later.")
		return
	}

	// If anonymous, never persist the display name; keep only the email (for
	// deduplication) and force contact_consent off.
	storedName := name
	contactConsent := req.ContactConsent
	if req.Anonymous {
		storedName = ""
		contactConsent = false
	}

	replaced, err := s.store.UpsertSubmission(ctx, store.SubmissionInput{
		Email:          email,
		Name:           storedName,
		Anonymous:      req.Anonymous,
		ContactConsent: contactConsent,
		IPHash:         ipHash,
		Ratings:        ratings,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save your submission.")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status": "created", "replaced": replaced,
	})
}

// handleContributors returns names of non-anonymous contributors. Disabled
// unless EXPOSE_CONTRIBUTORS=true. Never returns emails.
func (s *Server) handleContributors(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := withTimeout(r, 10*time.Second)
	defer cancel()
	names, err := s.store.ContributorNames(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to load contributors.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"contributors": names})
}
