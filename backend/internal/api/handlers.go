package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"ai-safety-atlas/backend/internal/store"
)

// handleHealth is also what the frontend pings to wake a cold-started instance.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "degraded",
			"detail": "no database configured; feedback submission is disabled",
		})
		return
	}
	ctx, cancel := contextWithTimeout(r, 3*time.Second)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		slog.Error("health ping failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleSummary returns aggregates only. Individual submissions are never
// readable through the public API.
func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeJSON(w, http.StatusOK, store.Summary{
			Iteration: s.cfg.Iteration,
			PerArea:   map[string]*store.RatingSummary{},
			PerAgenda: map[string]*store.RatingSummary{},
		})
		return
	}
	ctx, cancel := contextWithTimeout(r, 8*time.Second)
	defer cancel()

	sum, err := s.summary.Get(ctx)
	if err != nil {
		slog.Error("summary failed", "error", err)
		writeError(w, http.StatusServiceUnavailable, "summary temporarily unavailable")
		return
	}
	if sum.PerArea == nil {
		sum.PerArea = map[string]*store.RatingSummary{}
	}
	if sum.PerAgenda == nil {
		sum.PerAgenda = map[string]*store.RatingSummary{}
	}
	w.Header().Set("Cache-Control", "public, max-age=30")
	writeJSON(w, http.StatusOK, sum)
}

// handleFeedback is the only write path in the system.
func (s *Server) handleFeedback(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !s.limiter.Allow(ip) {
		w.Header().Set("Retry-After", "60")
		writeError(w, http.StatusTooManyRequests, "too many submissions, please wait a minute")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req FeedbackRequest
	if err := dec.Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "malformed JSON body")
		return
	}

	// Honeypot: answer as if it worked so a bot gets no signal to adapt to.
	if req.Website != "" {
		slog.Info("honeypot triggered")
		writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
		return
	}

	clean, verr := Validate(req)
	if verr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":  "validation failed",
			"field":  verr.Field,
			"detail": verr.Message,
		})
		return
	}

	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "feedback storage is not configured")
		return
	}

	ctx, cancel := contextWithTimeout(r, 8*time.Second)
	defer cancel()

	if s.turnstile != nil {
		if err := s.turnstile.Verify(ctx, req.TurnstileToken, ip); err != nil {
			slog.Warn("turnstile verification failed", "error", err)
			writeError(w, http.StatusForbidden, "bot check failed, please reload and try again")
			return
		}
	}

	err := s.store.InsertFeedback(ctx, store.Feedback{
		AgendaID:        clean.AgendaID,
		AreaTag:         clean.AreaTag,
		ProblemID:       clean.ProblemID,
		AreaMaturity:    clean.AreaMaturity,
		SubareaMaturity: clean.SubareaMaturity,
		Familiarity:     clean.Familiarity,
		Notes:           clean.Notes,
		SubmitterName:   clean.Name,
		SubmitterEmail:  clean.Email,
		IsAnonymous:     clean.IsAnonymous,
		ContactConsent:  clean.ContactConsent,
		ReuseConsent:    clean.ReuseConsent,
		SubmitterToken:  clean.SubmitterToken,
		Iteration:       s.cfg.Iteration,
		IPHash:          hashIP(s.cfg.IPHashSalt, ip),
		UserAgent:       truncateRunes(r.UserAgent(), MaxUARunes),
	})
	if err != nil {
		// The error may quote SQL or connection details; log it, do not return it.
		slog.Error("storing feedback failed", "agenda_id", clean.AgendaID, "area_tag", clean.AreaTag, "problem_id", clean.ProblemID, "error", err)
		writeError(w, http.StatusInternalServerError, "could not store your feedback, please try again")
		return
	}

	s.summary.Invalidate()
	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

// handleExport returns raw rows, including names and emails, to the operator.
func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if !bearerTokenOK(r.Header.Get("Authorization"), s.cfg.AdminToken) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="admin"`)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "no database configured")
		return
	}
	if f := r.URL.Query().Get("format"); f != "" && f != "json" {
		writeError(w, http.StatusBadRequest, "only format=json is supported")
		return
	}
	limit := 10000
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			writeError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		limit = n
	}

	ctx, cancel := contextWithTimeout(r, 20*time.Second)
	defer cancel()

	rows, err := s.store.Export(ctx, limit)
	if err != nil {
		slog.Error("export failed", "error", err)
		writeError(w, http.StatusInternalServerError, "export failed")
		return
	}
	// Downloaded, never rendered as a page: no HTML context, and the JSON encoder
	// escapes <, > and & regardless.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Disposition", `attachment; filename="feedback-export.json"`)
	writeJSON(w, http.StatusOK, map[string]any{"count": len(rows), "rows": rows})
}
