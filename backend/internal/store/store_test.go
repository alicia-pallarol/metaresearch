package store

import (
	"context"
	"os"
	"testing"
	"time"
)

// These tests exercise the real SQL, the migration, the insert, the aggregate and
// the export, against a real Postgres. They are skipped unless one is provided:
//
//	TEST_DATABASE_URL=postgres://... go test ./internal/store/
//
// A free Neon branch is the easiest way to run them (create a branch of the
// project, use its pooled connection string, delete the branch afterwards). CI
// can do the same. Everything they touch is inside a transaction-free test table
// that the migration creates, so run them against a scratch database, not
// production: the aggregate assertions assume they start empty.
func testStore(t *testing.T) *Store {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to a scratch Postgres to run the store integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	st, err := Open(ctx, url)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(st.Close)

	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("migrating: %v", err)
	}
	// Migrations must be safe to run twice: the operator may have pasted the SQL
	// into the Neon console before the service ever booted.
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("migrating a second time should be a no-op, got: %v", err)
	}

	if _, err := st.pool.Exec(ctx, `truncate feedback`); err != nil {
		t.Fatalf("clearing feedback: %v", err)
	}
	return st
}

func TestStoreRoundTrip(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()

	const token = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"
	rows := []Feedback{
		{AgendaID: "IN6", AreaTag: "IN", AreaMaturity: "Early / partial",
			SubareaMaturity: map[string]string{"IN6": "Untested"},
			Familiarity:     3, Notes: "solid", SubmitterToken: token, IsAnonymous: true, ReuseConsent: true},
		{AgendaID: "IN6", AreaTag: "IN", AreaMaturity: "Untested",
			SubareaMaturity: map[string]string{"IN6": "Contested", "IN2": "Untested"},
			Familiarity:     1, SubmitterToken: token, IsAnonymous: true},
		{AgendaID: "DS1", AreaTag: "DS", AreaMaturity: "Strong existence proof", Familiarity: 2,
			SubmitterName: "A Researcher", SubmitterEmail: "r@example.org", ContactConsent: true,
			SubmitterToken: "11111111-2222-3333-4444-555555555555"},
		// Cell-level feedback: no agenda, an area x problem target instead.
		{AreaTag: "IN", ProblemID: "P4", AreaMaturity: "Untested", Familiarity: 2,
			Notes: "missing an agenda here", SubmitterToken: token, IsAnonymous: true},
	}
	for _, f := range rows {
		if err := st.InsertFeedback(ctx, f); err != nil {
			t.Fatalf("inserting (agenda=%q area=%q): %v", f.AgendaID, f.AreaTag, err)
		}
	}

	sum, err := st.Summary(ctx, 0)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	if sum.Totals.TotalSubmissions != 4 {
		t.Errorf("total submissions = %d, want 4", sum.Totals.TotalSubmissions)
	}
	if sum.Totals.UniqueSubmitters != 2 {
		t.Errorf("unique submitters = %d, want 2 (three rows share a token)", sum.Totals.UniqueSubmitters)
	}

	// Areas count one vote per submission, whatever the form was opened from: the
	// two IN6 rows and the IN x P4 cell row all vote on IN.
	in := sum.PerArea["IN"]
	if in == nil {
		t.Fatal("IN missing from the per-area summary")
	}
	if in.Count != 3 {
		t.Errorf("IN count = %d, want 3", in.Count)
	}
	if in.Maturity["Untested"] != 2 || in.Maturity["Early / partial"] != 1 {
		t.Errorf("IN maturity distribution = %v", in.Maturity)
	}
	if in.Familiarity["1"] != 1 || in.Familiarity["2"] != 1 || in.Familiarity["3"] != 1 {
		t.Errorf("IN familiarity distribution = %v", in.Familiarity)
	}
	if in.AvgFamiliarity != 2 {
		t.Errorf("IN average familiarity = %v, want 2", in.AvgFamiliarity)
	}

	// Agendas count only the readers who rated that agenda, so DS1, voted on as
	// an area but never as a subarea, is absent rather than present with a count.
	if _, ok := sum.PerAgenda["DS1"]; ok {
		t.Error("DS1 got no subarea vote, so it must not appear in the per-agenda summary")
	}
	if _, ok := sum.PerAgenda[""]; ok {
		t.Error("per-agenda summary must not contain a blank agenda key")
	}

	in6 := sum.PerAgenda["IN6"]
	if in6 == nil {
		t.Fatal("IN6 missing from the per-agenda summary")
	}
	if in6.Count != 2 {
		t.Errorf("IN6 count = %d, want 2", in6.Count)
	}
	if in6.Maturity["Untested"] != 1 || in6.Maturity["Contested"] != 1 {
		t.Errorf("IN6 maturity distribution = %v", in6.Maturity)
	}
	if in6.AvgFamiliarity != 2 {
		t.Errorf("IN6 average familiarity = %v, want 2", in6.AvgFamiliarity)
	}
	if in6.Familiarity["1"] != 1 || in6.Familiarity["3"] != 1 || in6.Familiarity["0"] != 0 {
		t.Errorf("IN6 familiarity distribution = %v", in6.Familiarity)
	}
	if in2 := sum.PerAgenda["IN2"]; in2 == nil || in2.Count != 1 {
		t.Errorf("IN2 should carry the one subarea vote it was given, got %+v", in2)
	}

	exported, err := st.Export(ctx, 100)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if len(exported) != 4 {
		t.Fatalf("exported %d rows, want 4", len(exported))
	}

	// The cell row: null agenda, area + problem set.
	cell := exported[3]
	if cell.AgendaID != nil {
		t.Errorf("cell row agenda_id should be NULL, got %v", *cell.AgendaID)
	}
	if cell.AreaTag == nil || *cell.AreaTag != "IN" || cell.ProblemID == nil || *cell.ProblemID != "P4" {
		t.Errorf("cell row target = area %v, problem %v", cell.AreaTag, cell.ProblemID)
	}
	if cell.SubareaMaturity != nil {
		t.Errorf("cell row rated no agenda, so subarea_maturity should be NULL, got %v", cell.SubareaMaturity)
	}

	// The votes survive the round trip in the shape they were cast in.
	if exported[1].SubareaMaturity["IN6"] != "Contested" || exported[1].SubareaMaturity["IN2"] != "Untested" {
		t.Errorf("row 2 subarea votes = %v", exported[1].SubareaMaturity)
	}
	// Empty optional fields are stored as NULL, so "not answered" and "answered
	// with an empty string" stay distinguishable.
	if exported[1].Notes != nil {
		t.Errorf("row 2 notes should be NULL, got %v", *exported[1].Notes)
	}
	// The retired question is never written, on any row.
	for i, row := range exported {
		if row.AgreeWithRating != nil {
			t.Errorf("row %d wrote agree_with_rating (%q); the question is retired", i+1, *row.AgreeWithRating)
		}
	}
	if exported[2].SubmitterEmail == nil || *exported[2].SubmitterEmail != "r@example.org" {
		t.Errorf("row 3 email = %v", exported[2].SubmitterEmail)
	}
}

// The tier check constraint is the last line of defence behind API validation: a
// vote for a tier the map does not have must not reach the table.
func TestStoreRejectsUnknownTier(t *testing.T) {
	st := testStore(t)

	err := st.InsertFeedback(context.Background(), Feedback{
		AgendaID:       "EA1",
		AreaTag:        "EA",
		AreaMaturity:   "Extremely Robust",
		Familiarity:    2,
		SubmitterToken: "3f2504e0-4f89-11d3-9a0c-0305e82c3301",
		IsAnonymous:    true,
	})
	if err == nil {
		t.Fatal("expected the tier check constraint to reject an invented tier")
	}
}

func TestStoreRejectsOutOfRangeFamiliarity(t *testing.T) {
	st := testStore(t)

	// The check constraint is the last line of defence behind API validation.
	err := st.InsertFeedback(context.Background(), Feedback{
		AgendaID:       "EA1",
		AreaTag:        "EA",
		AreaMaturity:   "Untested",
		Familiarity:    9,
		SubmitterToken: "3f2504e0-4f89-11d3-9a0c-0305e82c3301",
		IsAnonymous:    true,
	})
	if err == nil {
		t.Fatal("expected the familiarity check constraint to reject 9")
	}
}
