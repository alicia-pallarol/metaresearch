// Package store owns every database access in the service. There is exactly one
// table (feedback) and exactly one write path (InsertFeedback).
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is a thin wrapper over a pgx pool. All statements are parameterised;
// no SQL is ever assembled from user input.
type Store struct {
	pool *pgxpool.Pool
}

// Feedback is one researcher submission, already validated by the API layer.
//
// The target says where the form was opened, an agenda (AgendaID set) or an
// Area x Problem cell (ProblemID set), and AreaTag is always populated, because
// the vote it carries is about the research area. The API guarantees the shape.
type Feedback struct {
	AgendaID  string // "" for cell-level feedback
	AreaTag   string // always set: the area the vote is about
	ProblemID string // "" for agenda-level feedback
	// AreaMaturity is the required vote: one tier for the area as a whole.
	AreaMaturity string
	// SubareaMaturity is agenda id -> tier; nil when no agenda was rated.
	SubareaMaturity map[string]string
	Familiarity     int16
	Notes           string
	SubmitterName   string
	SubmitterEmail  string
	IsAnonymous     bool
	ContactConsent  bool
	ReuseConsent    bool
	SubmitterToken  string // uuid
	Iteration       int
	IPHash          string
	UserAgent       string
}

// RatingSummary is the public aggregate for one research area or one agenda. It
// deliberately carries no identifying information, only counts.
//
// Both histograms are returned raw rather than reduced to an average, because the
// front end needs the whole shape: it computes the best/typical/worst-case bands
// from the maturity distribution, and a mean over six tiers, two of which are not
// rungs on the evidence ladder, would not mean anything anyway.
type RatingSummary struct {
	Count int `json:"count"`
	// Maturity is tier -> how many readers voted it.
	Maturity map[string]int `json:"maturity"`
	// Familiarity is "0".."3" -> how many readers reported it.
	Familiarity    map[string]int `json:"familiarity"`
	AvgFamiliarity float64        `json:"avg_familiarity"`
}

// Totals are the headline numbers shown in the hero.
type Totals struct {
	TotalSubmissions int `json:"total_submissions"`
	UniqueSubmitters int `json:"unique_submitters"`
}

// Summary is the whole of GET /api/summary.
//
// PerArea counts one vote per submission. PerAgenda counts only the submissions
// that rated that agenda specifically, so the two are different populations and an
// agenda's count is never inflated by people who only rated its area.
type Summary struct {
	Iteration int                       `json:"iteration"`
	Totals    Totals                    `json:"totals"`
	PerArea   map[string]*RatingSummary `json:"per_area"`
	PerAgenda map[string]*RatingSummary `json:"per_agenda"`
}

// ExportRow is a raw row for the admin export. It contains PII and is only ever
// returned behind the admin bearer token.
type ExportRow struct {
	ID        int64   `json:"id"`
	AgendaID  *string `json:"agenda_id"`
	AreaTag   *string `json:"area_tag"`
	ProblemID *string `json:"problem_id"`
	// AreaMaturity and SubareaMaturity are the votes; nil on rows written before
	// the map asked for them.
	AreaMaturity    *string           `json:"area_maturity"`
	SubareaMaturity map[string]string `json:"subarea_maturity"`
	Familiarity     int16             `json:"familiarity"`
	// AgreeWithRating is only ever set on rows from before the agree/disagree
	// question was retired. Nothing writes it; it is exported so the older
	// answers stay readable.
	AgreeWithRating *string   `json:"agree_with_rating"`
	Notes           *string   `json:"notes"`
	SubmitterName   *string   `json:"submitter_name"`
	SubmitterEmail  *string   `json:"submitter_email"`
	IsAnonymous     bool      `json:"is_anonymous"`
	ContactConsent  bool      `json:"contact_consent"`
	ReuseConsent    bool      `json:"reuse_consent"`
	SubmitterToken  string    `json:"submitter_token"`
	Iteration       int       `json:"iteration"`
	CreatedAt       time.Time `json:"created_at"`
}

// Open dials Postgres and verifies the connection.
func Open(ctx context.Context, databaseURL string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing DATABASE_URL: %w", err)
	}
	// Free-tier Postgres has a small connection budget and the service is tiny.
	cfg.MaxConns = 4
	cfg.MaxConnIdleTime = 2 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connecting to Postgres: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging Postgres: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

const insertFeedbackSQL = `
insert into feedback (
  agenda_id, area_tag, problem_id, area_maturity, subarea_maturity,
  familiarity, notes,
  submitter_name, submitter_email, is_anonymous,
  contact_consent, reuse_consent, submitter_token,
  iteration, ip_hash, user_agent
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`

// InsertFeedback stores one submission. Empty optional strings are stored as
// NULL so "not given" and "given as empty" do not blur together, and so the
// agenda-vs-cell target columns stay cleanly one-or-the-other.
//
// agree_with_rating is not written: the question is retired. The column still
// holds the answers given while it was asked.
func (s *Store) InsertFeedback(ctx context.Context, f Feedback) error {
	subareas, err := marshalSubareas(f.SubareaMaturity)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, insertFeedbackSQL,
		nullable(f.AgendaID),
		nullable(f.AreaTag),
		nullable(f.ProblemID),
		nullable(f.AreaMaturity),
		subareas,
		f.Familiarity,
		nullable(f.Notes),
		nullable(f.SubmitterName),
		nullable(f.SubmitterEmail),
		f.IsAnonymous,
		f.ContactConsent,
		f.ReuseConsent,
		f.SubmitterToken,
		f.Iteration,
		nullable(f.IPHash),
		nullable(f.UserAgent),
	)
	if err != nil {
		return fmt.Errorf("inserting feedback: %w", err)
	}
	return nil
}

// marshalSubareas renders the subarea votes as jsonb, or NULL when there are
// none: an empty object and "rated no agenda" are the same claim, and only one of
// them should be storable.
func marshalSubareas(votes map[string]string) ([]byte, error) {
	if len(votes) == 0 {
		return nil, nil
	}
	raw, err := json.Marshal(votes)
	if err != nil {
		return nil, fmt.Errorf("encoding subarea votes: %w", err)
	}
	return raw, nil
}

// Both aggregate queries return one row per (subject, tier, familiarity) triple
// and are folded into histograms in Go. There are at most a few hundred rows
// either way, and it keeps the tier vocabulary out of the SQL: a renamed tier is
// then a data change, not a schema change.
const summaryPerAreaSQL = `
select area_tag, area_maturity, familiarity, count(*)
from feedback
where area_tag is not null and area_maturity is not null
group by area_tag, area_maturity, familiarity`

// Agendas are aggregated from the subarea votes, so an agenda's count is the
// number of readers who rated that agenda, not the number who rated its area.
const summaryPerAgendaSQL = `
select kv.key, kv.value, f.familiarity, count(*)
from feedback f, lateral jsonb_each_text(f.subarea_maturity) kv
where f.subarea_maturity is not null
group by kv.key, kv.value, f.familiarity`

const summaryTotalsSQL = `
select count(*), count(distinct submitter_token) from feedback`

// Summary aggregates the feedback table. Aggregates run across every iteration on
// purpose: a researcher's familiarity with an agenda does not expire when the map
// is re-cut, and the contributor counter should not drop to zero on a new iteration.
func (s *Store) Summary(ctx context.Context, iteration int) (Summary, error) {
	out := Summary{
		Iteration: iteration,
		PerArea:   map[string]*RatingSummary{},
		PerAgenda: map[string]*RatingSummary{},
	}

	if err := s.foldRatings(ctx, summaryPerAreaSQL, out.PerArea); err != nil {
		return out, fmt.Errorf("querying per-area summary: %w", err)
	}
	if err := s.foldRatings(ctx, summaryPerAgendaSQL, out.PerAgenda); err != nil {
		return out, fmt.Errorf("querying per-agenda summary: %w", err)
	}

	if err := s.pool.QueryRow(ctx, summaryTotalsSQL).
		Scan(&out.Totals.TotalSubmissions, &out.Totals.UniqueSubmitters); err != nil {
		return out, fmt.Errorf("querying totals: %w", err)
	}
	return out, nil
}

// foldRatings runs a (subject, tier, familiarity, n) query and folds it into one
// RatingSummary per subject.
func (s *Store) foldRatings(ctx context.Context, sql string, into map[string]*RatingSummary) error {
	rows, err := s.pool.Query(ctx, sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	famTotal := map[string]int{}
	for rows.Next() {
		var (
			subject     string
			tier        string
			familiarity int16
			n           int
		)
		if err := rows.Scan(&subject, &tier, &familiarity, &n); err != nil {
			return fmt.Errorf("scanning summary row: %w", err)
		}
		sum := into[subject]
		if sum == nil {
			sum = &RatingSummary{Maturity: map[string]int{}, Familiarity: map[string]int{}}
			into[subject] = sum
		}
		sum.Count += n
		sum.Maturity[tier] += n
		sum.Familiarity[strconv.Itoa(int(familiarity))] += n
		famTotal[subject] += int(familiarity) * n
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading summary rows: %w", err)
	}

	for subject, sum := range into {
		if sum.Count > 0 {
			sum.AvgFamiliarity = round1(float64(famTotal[subject]) / float64(sum.Count))
		}
	}
	return nil
}

const exportSQL = `
select id, agenda_id, area_tag, problem_id, area_maturity, subarea_maturity,
       familiarity, agree_with_rating, notes,
       submitter_name, submitter_email, is_anonymous,
       contact_consent, reuse_consent, submitter_token, iteration, created_at
from feedback
order by id
limit $1`

// Export returns raw rows for the operator. Admin-only; never reachable publicly.
func (s *Store) Export(ctx context.Context, limit int) ([]ExportRow, error) {
	if limit <= 0 || limit > 10000 {
		limit = 10000
	}
	rows, err := s.pool.Query(ctx, exportSQL, limit)
	if err != nil {
		return nil, fmt.Errorf("querying export: %w", err)
	}
	defer rows.Close()

	out := make([]ExportRow, 0, 64)
	for rows.Next() {
		var (
			r        ExportRow
			subareas []byte
		)
		if err := rows.Scan(&r.ID, &r.AgendaID, &r.AreaTag, &r.ProblemID, &r.AreaMaturity, &subareas,
			&r.Familiarity, &r.AgreeWithRating, &r.Notes,
			&r.SubmitterName, &r.SubmitterEmail, &r.IsAnonymous, &r.ContactConsent,
			&r.ReuseConsent, &r.SubmitterToken, &r.Iteration, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning export row: %w", err)
		}
		if len(subareas) > 0 {
			if err := json.Unmarshal(subareas, &r.SubareaMaturity); err != nil {
				return nil, fmt.Errorf("decoding subarea votes on row %d: %w", r.ID, err)
			}
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading export rows: %w", err)
	}
	return out, nil
}

// Ping is used by the health endpoint.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func round1(f float64) float64 {
	return float64(int64(f*10+0.5)) / 10
}
