// Package config loads runtime configuration from the environment.
// No secret is ever hard-coded; every value below comes from an env var.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// DatabaseURL is the Postgres connection string (Neon pooled URL in production).
	// May be empty: the service then runs in degraded mode (map still served by the
	// frontend, feedback writes rejected with 503) which is handy for frontend dev.
	DatabaseURL string
	// AllowedOrigin is the exact frontend origin allowed by CORS, e.g.
	// https://ai-safety-atlas.pages.dev. "*" is rejected on purpose.
	AllowedOrigin string
	// AdminToken guards GET /api/admin/export. Empty disables the endpoint.
	AdminToken string
	Port       string
	// TurnstileSecret enables Cloudflare Turnstile verification when set.
	TurnstileSecret string
	// IPHashSalt salts the sha256 of the client IP. When empty a random salt is
	// generated at boot, which keeps hashes unlinkable across restarts.
	IPHashSalt string
	// Iteration is stamped on every stored row so feedback stays attached to the
	// version of the map it was given against.
	Iteration int
}

// Load reads the environment and validates it. It returns a human-readable error
// rather than panicking so the operator sees what is missing in the deploy logs.
func Load() (Config, error) {
	c := Config{
		DatabaseURL:     strings.TrimSpace(os.Getenv("DATABASE_URL")),
		AllowedOrigin:   strings.TrimSpace(os.Getenv("ALLOWED_ORIGIN")),
		AdminToken:      strings.TrimSpace(os.Getenv("ADMIN_TOKEN")),
		Port:            strings.TrimSpace(os.Getenv("PORT")),
		TurnstileSecret: strings.TrimSpace(os.Getenv("TURNSTILE_SECRET")),
		IPHashSalt:      strings.TrimSpace(os.Getenv("IP_HASH_SALT")),
	}

	if c.Port == "" {
		c.Port = "8080"
	}
	if _, err := strconv.Atoi(c.Port); err != nil {
		return c, fmt.Errorf("PORT must be a number, got %q", c.Port)
	}

	if c.AllowedOrigin == "" {
		return c, fmt.Errorf("ALLOWED_ORIGIN is required (exact frontend origin, e.g. https://example.pages.dev)")
	}
	if c.AllowedOrigin == "*" {
		return c, fmt.Errorf("ALLOWED_ORIGIN must be an exact origin, not \"*\"")
	}
	if !strings.HasPrefix(c.AllowedOrigin, "http://") && !strings.HasPrefix(c.AllowedOrigin, "https://") {
		return c, fmt.Errorf("ALLOWED_ORIGIN must include the scheme, got %q", c.AllowedOrigin)
	}
	c.AllowedOrigin = strings.TrimSuffix(c.AllowedOrigin, "/")

	if raw := strings.TrimSpace(os.Getenv("ITERATION")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return c, fmt.Errorf("ITERATION must be a non-negative integer, got %q", raw)
		}
		c.Iteration = n
	}

	if c.IPHashSalt == "" {
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			return c, fmt.Errorf("generating IP hash salt: %w", err)
		}
		c.IPHashSalt = hex.EncodeToString(buf)
	}

	return c, nil
}

// Degraded reports whether the service is running without a database.
func (c Config) Degraded() bool { return c.DatabaseURL == "" }
