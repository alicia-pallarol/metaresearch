// Package api wires the four HTTP endpoints of the service. There is exactly one
// write path (POST /api/feedback); everything else is read-only or aggregate.
package api

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"ai-safety-atlas/backend/internal/config"
	"ai-safety-atlas/backend/internal/store"
)

// Storage is what the API needs from the persistence layer.
type Storage interface {
	InsertFeedback(ctx context.Context, f store.Feedback) error
	Summary(ctx context.Context, iteration int) (store.Summary, error)
	Export(ctx context.Context, limit int) ([]store.ExportRow, error)
	Ping(ctx context.Context) error
}

// Server holds the dependencies of the HTTP layer.
type Server struct {
	cfg       config.Config
	store     Storage // nil in degraded mode (no DATABASE_URL)
	limiter   *RateLimiter
	summary   *summaryCache
	turnstile *turnstileVerifier
}

const summaryTTL = 45 * time.Second

// New builds a Server. A nil store is allowed and puts the service in degraded
// mode: reads answer with empty aggregates, writes answer 503.
func New(cfg config.Config, st Storage) *Server {
	s := &Server{
		cfg:     cfg,
		store:   st,
		limiter: NewRateLimiter(5, 30),
	}
	if st != nil {
		s.summary = newSummaryCache(st, cfg.Iteration, summaryTTL)
	}
	if cfg.TurnstileSecret != "" {
		s.turnstile = newTurnstileVerifier(cfg.TurnstileSecret)
	}
	return s
}

// Handler returns the routed, middleware-wrapped handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/summary", s.handleSummary)
	mux.HandleFunc("POST /api/feedback", s.handleFeedback)
	mux.HandleFunc("GET /api/admin/export", s.handleExport)

	return s.recoverPanics(s.cors(s.logRequests(mux)))
}

// cors answers preflights and stamps the single allowed origin. Anything else
// gets no CORS headers at all, so the browser blocks it.
func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == s.cfg.AllowedOrigin {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", s.cfg.AllowedOrigin)
			h.Set("Vary", "Origin")
			h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			h.Set("Access-Control-Max-Age", "86400")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		// Deliberately no query string, no body, no email, no notes: nothing a
		// submission carries should ever reach the logs.
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func (s *Server) recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "path", r.URL.Path, "panic", rec)
				writeError(w, http.StatusInternalServerError, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// writeJSON emits a JSON body. Struct fields carry no HTML, and the encoder
// escapes <, > and & by default, so feedback text cannot break out of the JSON.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writing response", "error", err)
	}
}

// writeError returns a structured error. Internal detail never leaves the process.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// clientIP prefers the left-most X-Forwarded-For entry, which is what Render's
// proxy sets. It is only ever used salted-and-hashed, or as a rate-limit key.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first, _, ok := strings.Cut(xff, ","); ok {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// hashIP is a salted sha256. The raw address is never stored or logged.
func hashIP(salt, ip string) string {
	if ip == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(salt + "|" + ip))
	return hex.EncodeToString(sum[:])
}

// bearerTokenOK compares in constant time so the admin token cannot be probed
// byte by byte.
func bearerTokenOK(header, expected string) bool {
	if expected == "" {
		return false
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	got := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return subtle.ConstantTimeCompare([]byte(got), []byte(expected)) == 1
}
