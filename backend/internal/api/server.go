package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/saige/backend/internal/config"
	"github.com/saige/backend/internal/store"
)

// Server holds handler dependencies.
type Server struct {
	cfg     *config.Config
	store   *store.Store
	limiter *rateLimiter
	origins map[string]bool
}

// NewServer builds the HTTP handler graph.
func NewServer(cfg *config.Config, st *store.Store) http.Handler {
	s := &Server{
		cfg:     cfg,
		store:   st,
		limiter: newRateLimiter(cfg.RateLimitPerHour, time.Hour),
		origins: map[string]bool{},
	}
	for _, o := range cfg.AllowedOrigins {
		s.origins[o] = true
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(s.securityHeaders)
	r.Use(s.cors)

	r.Get("/api/health", s.handleHealth)
	r.Get("/api/research-areas", s.handleResearchAreas)
	r.Post("/api/submissions", s.handleSubmit)
	if cfg.ExposeContributors {
		r.Get("/api/contributors", s.handleContributors)
	}

	return r
}

// securityHeaders sets conservative headers on every API response. The API
// serves JSON only, so a very strict CSP is appropriate.
func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Cross-Origin-Resource-Policy", "same-site")
		next.ServeHTTP(w, r)
	})
}

// cors permits only the configured origins and echoes them back explicitly.
func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := s.origins["*"] || s.origins[origin]
		if allowed && origin != "" {
			if s.origins["*"] {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type errorResponse struct {
	Error  string            `json:"error"`
	Fields []validationError `json:"fields,omitempty"`
}

func writeError(w http.ResponseWriter, status int, msg string, fields ...validationError) {
	writeJSON(w, status, errorResponse{Error: msg, Fields: fields})
}

// hashIP returns an HMAC-SHA256 of the client IP salted with the configured
// secret. The raw IP is never stored or logged.
func (s *Server) hashIP(r *http.Request) string {
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.IPHashSalt))
	mac.Write([]byte(strings.ToLower(ip)))
	return hex.EncodeToString(mac.Sum(nil))
}

// withTimeout derives a request-scoped context with a hard cap.
func withTimeout(r *http.Request, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), d)
}
