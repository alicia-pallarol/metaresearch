package store

import (
	"context"
	"os"
	"testing"
)

// testStore connects to TEST_DATABASE_URL and applies migrations. Integration
// tests skip cleanly when no test database is configured.
func testStore(t *testing.T) *Store {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run store integration tests")
	}
	ctx := context.Background()
	st, err := New(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(st.Close)

	// The migrations live at the repo root, two levels up from this package.
	if err := st.Migrate(ctx, "../../../migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// Start each test from a clean submissions state.
	if _, err := st.pool.Exec(ctx, `DELETE FROM submissions`); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	return st
}

// TestAggregationSmoke is the required smoke test for the aggregation query:
// it inserts two submissions and verifies per-area n, average and distribution.
func TestAggregationSmoke(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()

	tagToID, err := st.TagToID(ctx)
	if err != nil {
		t.Fatalf("tag map: %v", err)
	}
	ea, in := tagToID["EA"], tagToID["IN"]

	if _, err := st.UpsertSubmission(ctx, SubmissionInput{
		Email: "a@example.com", Anonymous: true,
		Ratings: map[int]int{ea: 3, in: 1},
	}); err != nil {
		t.Fatalf("upsert a: %v", err)
	}
	if _, err := st.UpsertSubmission(ctx, SubmissionInput{
		Email: "b@example.com", Anonymous: true,
		Ratings: map[int]int{ea: 1}, // leaves IN blank
	}); err != nil {
		t.Fatalf("upsert b: %v", err)
	}

	agg, err := st.AreaAggregates(ctx, 3)
	if err != nil {
		t.Fatalf("aggregates: %v", err)
	}
	if agg.TotalSubmissions != 2 {
		t.Errorf("total submissions = %d, want 2", agg.TotalSubmissions)
	}

	byTag := map[string]ResearchArea{}
	for _, a := range agg.Areas {
		byTag[a.Tag] = a
	}
	if len(agg.Areas) != 12 {
		t.Errorf("expected 12 areas, got %d", len(agg.Areas))
	}
	if got := byTag["EA"]; got.N != 2 || got.Average == nil || *got.Average != 2.0 {
		t.Errorf("EA n=%d avg=%v, want n=2 avg=2.0", got.N, got.Average)
	}
	if got := byTag["EA"].Distribution; got["1"] != 1 || got["3"] != 1 {
		t.Errorf("EA distribution = %v, want one 1 and one 3", got)
	}
	if got := byTag["IN"]; got.N != 1 || got.Average == nil || *got.Average != 1.0 {
		t.Errorf("IN n=%d avg=%v, want n=1 avg=1.0", got.N, got.Average)
	}
	if got := byTag["SA"]; got.N != 0 || got.Average != nil {
		t.Errorf("SA should have no ratings, got n=%d avg=%v", got.N, got.Average)
	}
}

// TestUpsertReplacesByEmail verifies the dedup rule: resubmitting from the same
// email replaces the previous submission and its ratings.
func TestUpsertReplacesByEmail(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()
	tagToID, _ := st.TagToID(ctx)
	ea := tagToID["EA"]

	if replaced, err := st.UpsertSubmission(ctx, SubmissionInput{
		Email: "dup@example.com", Anonymous: true, Ratings: map[int]int{ea: 0},
	}); err != nil || replaced {
		t.Fatalf("first upsert replaced=%v err=%v, want replaced=false", replaced, err)
	}

	replaced, err := st.UpsertSubmission(ctx, SubmissionInput{
		Email: "dup@example.com", Anonymous: true, Ratings: map[int]int{ea: 3},
	})
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	if !replaced {
		t.Errorf("second upsert should report replaced=true")
	}

	agg, _ := st.AreaAggregates(ctx, 3)
	if agg.TotalSubmissions != 1 {
		t.Errorf("total submissions = %d, want 1 (dedup)", agg.TotalSubmissions)
	}
	for _, a := range agg.Areas {
		if a.Tag == "EA" {
			if a.N != 1 || a.Average == nil || *a.Average != 3.0 {
				t.Errorf("EA after replace n=%d avg=%v, want n=1 avg=3.0", a.N, a.Average)
			}
		}
	}
}
