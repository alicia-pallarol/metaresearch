// Package store contains the Postgres data access layer. All SQL is
// parameterized; no query interpolates user input into the statement text.
package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps a pgx connection pool.
type Store struct {
	pool *pgxpool.Pool
}

// New opens a pooled connection to Postgres and verifies connectivity.
func New(ctx context.Context, databaseURL string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 5 // Neon free tier has a modest connection limit.
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close releases the underlying pool.
func (s *Store) Close() { s.pool.Close() }

// Ping checks database connectivity for the health endpoint.
func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

// ResearchArea is a single row of the framework, enriched with aggregate
// familiarity statistics computed across all submissions.
type ResearchArea struct {
	ID           int            `json:"id"`
	ProblemGroup string         `json:"problem_group"`
	Tag          string         `json:"tag"`
	Name         string         `json:"name"`
	Definition   string         `json:"definition"`
	N            int            `json:"n"`            // number of ratings for this area
	Average      *float64       `json:"average"`      // null when n == 0
	Distribution map[string]int `json:"distribution"` // "0".."3" -> count
}

// Aggregates is the public payload powering the table and heatmap.
type Aggregates struct {
	TotalSubmissions   int            `json:"total_submissions"`
	LowSampleThreshold int            `json:"low_sample_threshold"`
	Areas              []ResearchArea `json:"areas"`
}

// AreaAggregates returns every research area together with its rating stats.
func (s *Store) AreaAggregates(ctx context.Context, lowSample int) (*Aggregates, error) {
	const q = `
SELECT ra.id, ra.problem_group, ra.tag, ra.name, ra.definition,
       COALESCE(COUNT(r.familiarity), 0)                                  AS n,
       AVG(r.familiarity)                                                 AS average,
       COUNT(*) FILTER (WHERE r.familiarity = 0)                          AS d0,
       COUNT(*) FILTER (WHERE r.familiarity = 1)                          AS d1,
       COUNT(*) FILTER (WHERE r.familiarity = 2)                          AS d2,
       COUNT(*) FILTER (WHERE r.familiarity = 3)                          AS d3
FROM research_areas ra
LEFT JOIN ratings r ON r.research_area_id = ra.id
GROUP BY ra.id
ORDER BY ra.sort_order, ra.id`

	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query aggregates: %w", err)
	}
	defer rows.Close()

	out := &Aggregates{LowSampleThreshold: lowSample, Areas: []ResearchArea{}}
	for rows.Next() {
		var a ResearchArea
		var avg *float64
		var d0, d1, d2, d3 int
		if err := rows.Scan(&a.ID, &a.ProblemGroup, &a.Tag, &a.Name, &a.Definition,
			&a.N, &avg, &d0, &d1, &d2, &d3); err != nil {
			return nil, fmt.Errorf("scan aggregate: %w", err)
		}
		a.Average = avg
		a.Distribution = map[string]int{"0": d0, "1": d1, "2": d2, "3": d3}
		out.Areas = append(out.Areas, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM submissions`).
		Scan(&out.TotalSubmissions); err != nil {
		return nil, fmt.Errorf("count submissions: %w", err)
	}
	return out, nil
}

// TagToID returns a map of research-area tag -> id for validating submissions.
func (s *Store) TagToID(ctx context.Context) (map[string]int, error) {
	rows, err := s.pool.Query(ctx, `SELECT tag, id FROM research_areas`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]int)
	for rows.Next() {
		var tag string
		var id int
		if err := rows.Scan(&tag, &id); err != nil {
			return nil, err
		}
		m[tag] = id
	}
	return m, rows.Err()
}

// SubmissionInput is a validated submission ready to persist.
type SubmissionInput struct {
	Email          string
	Name           string
	Anonymous      bool
	ContactConsent bool
	IPHash         string
	// Ratings maps research_area_id -> familiarity (0..3).
	Ratings map[int]int
}

// UpsertSubmission replaces any prior submission from the same email and stores
// its ratings atomically. Returns true if it replaced an existing submission.
func (s *Store) UpsertSubmission(ctx context.Context, in SubmissionInput) (replaced bool, err error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id int64
	err = tx.QueryRow(ctx, `
INSERT INTO submissions (email, name, anonymous, contact_consent, ip_hash, updated_at)
VALUES ($1, $2, $3, $4, $5, now())
ON CONFLICT (email) DO UPDATE SET
    name            = EXCLUDED.name,
    anonymous       = EXCLUDED.anonymous,
    contact_consent = EXCLUDED.contact_consent,
    ip_hash         = EXCLUDED.ip_hash,
    updated_at      = now()
RETURNING id, (xmax <> 0) AS existed`,
		in.Email, in.Name, in.Anonymous, in.ContactConsent, in.IPHash).
		Scan(&id, &replaced)
	if err != nil {
		return false, fmt.Errorf("upsert submission: %w", err)
	}

	// Clear any prior ratings, then insert the new set.
	if _, err = tx.Exec(ctx, `DELETE FROM ratings WHERE submission_id = $1`, id); err != nil {
		return false, fmt.Errorf("clear ratings: %w", err)
	}
	for areaID, fam := range in.Ratings {
		if _, err = tx.Exec(ctx,
			`INSERT INTO ratings (submission_id, research_area_id, familiarity) VALUES ($1, $2, $3)`,
			id, areaID, fam); err != nil {
			return false, fmt.Errorf("insert rating: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit: %w", err)
	}
	return replaced, nil
}

// ContributorNames returns the display names of non-anonymous submissions.
// Emails are never returned. Used only when EXPOSE_CONTRIBUTORS is enabled.
func (s *Store) ContributorNames(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
SELECT name FROM submissions
WHERE anonymous = FALSE AND name <> ''
ORDER BY lower(name)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	names := []string{}
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}

// CountRecentByIPHash returns how many submissions the given IP hash has made
// since the provided time; used for rate limiting.
func (s *Store) CountRecentByIPHash(ctx context.Context, ipHash string, since time.Time) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM submissions WHERE ip_hash = $1 AND updated_at >= $2`,
		ipHash, since).Scan(&n)
	return n, err
}

// ErrNoRows is re-exported for callers that need to detect empty lookups.
var ErrNoRows = pgx.ErrNoRows

// IsNoRows reports whether err is a "no rows" error.
func IsNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }
