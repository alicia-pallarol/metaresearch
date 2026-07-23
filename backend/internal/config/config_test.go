package config

import "testing"

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
		check   func(*testing.T, Config)
	}{
		{
			name: "minimal valid configuration",
			env:  map[string]string{"ALLOWED_ORIGIN": "https://atlas.pages.dev"},
			check: func(t *testing.T, c Config) {
				if c.Port != "8080" {
					t.Errorf("Port = %q, want the 8080 default", c.Port)
				}
				if !c.Degraded() {
					t.Error("no DATABASE_URL should mean degraded mode")
				}
				if c.IPHashSalt == "" {
					t.Error("a salt should be generated when none is configured")
				}
			},
		},
		{
			name:    "origin is required",
			env:     map[string]string{},
			wantErr: true,
		},
		{
			name:    "wildcard origin is refused",
			env:     map[string]string{"ALLOWED_ORIGIN": "*"},
			wantErr: true,
		},
		{
			name:    "origin needs a scheme",
			env:     map[string]string{"ALLOWED_ORIGIN": "atlas.pages.dev"},
			wantErr: true,
		},
		{
			name: "trailing slash is trimmed so CORS compares equal",
			env:  map[string]string{"ALLOWED_ORIGIN": "https://atlas.pages.dev/"},
			check: func(t *testing.T, c Config) {
				if c.AllowedOrigin != "https://atlas.pages.dev" {
					t.Errorf("AllowedOrigin = %q", c.AllowedOrigin)
				}
			},
		},
		{
			name:    "port must be numeric",
			env:     map[string]string{"ALLOWED_ORIGIN": "https://x.dev", "PORT": "http"},
			wantErr: true,
		},
		{
			name:    "iteration must be a non-negative integer",
			env:     map[string]string{"ALLOWED_ORIGIN": "https://x.dev", "ITERATION": "-1"},
			wantErr: true,
		},
		{
			name: "full configuration",
			env: map[string]string{
				"ALLOWED_ORIGIN": "https://atlas.pages.dev",
				"DATABASE_URL":   "postgres://user:pw@host/db",
				"ADMIN_TOKEN":    "token",
				"PORT":           "9000",
				"ITERATION":      "1",
				"IP_HASH_SALT":   "pepper",
			},
			check: func(t *testing.T, c Config) {
				if c.Degraded() {
					t.Error("a DATABASE_URL should leave degraded mode")
				}
				if c.Iteration != 1 || c.Port != "9000" || c.IPHashSalt != "pepper" {
					t.Errorf("unexpected config: %+v", c)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{
				"ALLOWED_ORIGIN", "DATABASE_URL", "ADMIN_TOKEN",
				"PORT", "ITERATION", "IP_HASH_SALT", "TURNSTILE_SECRET",
			} {
				t.Setenv(key, "")
			}
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			cfg, err := Load()
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, cfg)
			}
		})
	}
}

func TestLoadGeneratesDistinctSalts(t *testing.T) {
	t.Setenv("ALLOWED_ORIGIN", "https://atlas.pages.dev")
	t.Setenv("IP_HASH_SALT", "")

	a, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.IPHashSalt == b.IPHashSalt {
		t.Error("generated salts should differ between boots")
	}
}
