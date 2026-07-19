package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Migrate applies any not-yet-applied `*.up.sql` migrations found in dir, in
// filename order, inside a tracking table. It is safe to run on every startup.
//
// If dir is empty, a small set of conventional locations is probed so the
// binary works whether it is launched from the repo root or the backend dir.
func (s *Store) Migrate(ctx context.Context, dir string) error {
	resolved, err := resolveMigrationsDir(dir)
	if err != nil {
		return err
	}

	if _, err := s.pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := map[string]bool{}
	rows, err := s.pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	entries, err := os.ReadDir(resolved)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", resolved, err)
	}
	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, name := range upFiles {
		version := strings.TrimSuffix(name, ".up.sql")
		if applied[version] {
			continue
		}
		sqlBytes, err := os.ReadFile(filepath.Join(resolved, name))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		// Each migration runs in its own transaction.
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return err
		}
		// Migration files contain multiple semicolon-separated statements, which
		// pgx's default (extended) protocol rejects in a single Exec. The simple
		// protocol permits multi-statement command strings.
		if _, err := tx.Exec(ctx, string(sqlBytes), pgx.QueryExecModeSimpleProtocol); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record %s: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", name, err)
		}
	}
	return nil
}

// resolveMigrationsDir returns dir if it exists, otherwise probes conventional
// locations relative to the working directory and the executable.
func resolveMigrationsDir(dir string) (string, error) {
	candidates := []string{}
	if dir != "" {
		candidates = append(candidates, dir)
	}
	candidates = append(candidates, "migrations", "../migrations", "./backend/migrations")
	if exe, err := os.Executable(); err == nil {
		base := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(base, "migrations"),
			filepath.Join(base, "..", "migrations"))
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("could not locate migrations directory (tried %v); set MIGRATIONS_PATH", candidates)
}
