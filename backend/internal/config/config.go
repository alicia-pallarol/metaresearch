// Package config loads runtime configuration exclusively from environment
// variables. No secrets are ever hard-coded; see .env.example for the contract.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for the API server.
type Config struct {
	// DatabaseURL is the Postgres connection string (e.g. from Neon).
	DatabaseURL string
	// Port is the TCP port to listen on. Render provides this via $PORT.
	Port string
	// AllowedOrigins is the set of exact origins permitted by CORS.
	// Comma-separated in the env var; "*" is accepted for local dev only.
	AllowedOrigins []string
	// IPHashSalt salts the per-IP hash used only for rate limiting.
	IPHashSalt string
	// MigrationsPath points at the directory of .sql migration files.
	MigrationsPath string
	// LowSampleThreshold: below this many ratings an area is flagged low-sample.
	LowSampleThreshold int
	// RateLimitPerHour caps submissions per hashed IP per rolling hour.
	RateLimitPerHour int
	// ExposeContributors, when true, allows the API to return the names of
	// non-anonymous contributors. Defaults to false (aggregates only).
	ExposeContributors bool
}

// Load reads configuration from the environment, applying sane defaults and
// validating that the required values are present.
func Load() (*Config, error) {
	c := &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		Port:               getenv("PORT", "8080"),
		IPHashSalt:         os.Getenv("IP_HASH_SALT"),
		MigrationsPath:     getenv("MIGRATIONS_PATH", ""),
		LowSampleThreshold: getenvInt("LOW_SAMPLE_THRESHOLD", 3),
		RateLimitPerHour:   getenvInt("RATE_LIMIT_PER_HOUR", 5),
		ExposeContributors: getenvBool("EXPOSE_CONTRIBUTORS", false),
	}

	origin := getenv("ALLOWED_ORIGIN", "http://localhost:5173")
	for _, o := range strings.Split(origin, ",") {
		if o = strings.TrimSpace(o); o != "" {
			c.AllowedOrigins = append(c.AllowedOrigins, o)
		}
	}

	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if c.IPHashSalt == "" {
		// A missing salt would make IP hashes trivially reversible; refuse to run.
		return nil, fmt.Errorf("IP_HASH_SALT is required (use a long random string)")
	}
	if len(c.AllowedOrigins) == 0 {
		return nil, fmt.Errorf("ALLOWED_ORIGIN is required")
	}
	return c, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getenvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
