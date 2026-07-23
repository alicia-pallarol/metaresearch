package atlas

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// atlasPath is the dataset the whole project is built from. The file lives
// outside the Go module, so it is read at test time rather than embedded: the
// binary only ever needs the id list in ids.go.
const atlasPath = "../../../data/atlas.json"

func TestKnownAgendaID(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"EA1", true},
		{"TG3", true},
		{"IN6", true},
		{"", false},
		{"ea1", false}, // ids are case-sensitive
		{"ZZ9", false},
		{"EA1; drop table feedback", false},
	}
	for _, tc := range tests {
		if got := KnownAgendaID(tc.id); got != tc.want {
			t.Errorf("KnownAgendaID(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestKnownAreaTag(t *testing.T) {
	tests := []struct {
		tag  string
		want bool
	}{
		{"EA", true},
		{"IN", true},
		{"TG", true},
		{"", false},
		{"ea", false},
		{"EA1", false}, // an agenda id is not an area tag
		{"ZZ", false},
	}
	for _, tc := range tests {
		if got := KnownAreaTag(tc.tag); got != tc.want {
			t.Errorf("KnownAreaTag(%q) = %v, want %v", tc.tag, got, tc.want)
		}
	}
}

func TestKnownProblemID(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"P1", true},
		{"P12", true},
		{"", false},
		{"p1", false},
		{"P13", false},
		{"P0", false},
	}
	for _, tc := range tests {
		if got := KnownProblemID(tc.id); got != tc.want {
			t.Errorf("KnownProblemID(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestKnownTier(t *testing.T) {
	tests := []struct {
		tier string
		want bool
	}{
		{"Robust (small scale)", true},
		{"Contested", true},
		{"Never demonstrated", true},
		{"", false},
		{"robust (small scale)", false}, // tiers are case-sensitive
		{"Robust", false},               // and must be spelled in full
		{"(blank)", false},              // blank is not a tier anyone can vote for
		{"Excellent", false},
	}
	for _, tc := range tests {
		if got := KnownTier(tc.tier); got != tc.want {
			t.Errorf("KnownTier(%q) = %v, want %v", tc.tier, got, tc.want)
		}
	}
}

func TestAreaOfAgenda(t *testing.T) {
	tests := []struct {
		id      string
		wantTag string
		wantOK  bool
	}{
		{"EA1", "EA", true},
		{"IN6", "IN", true},
		{"TG3", "TG", true},
		{"ZZ9", "", false},
		{"", "", false},
	}
	for _, tc := range tests {
		tag, ok := AreaOfAgenda(tc.id)
		if tag != tc.wantTag || ok != tc.wantOK {
			t.Errorf("AreaOfAgenda(%q) = %q, %v; want %q, %v", tc.id, tag, ok, tc.wantTag, tc.wantOK)
		}
	}
}

func TestIDCount(t *testing.T) {
	if len(Iteration0AgendaIDs) != 58 {
		t.Fatalf("the curated map has 58 agendas, ids.go carries %d", len(Iteration0AgendaIDs))
	}
	if len(Iteration0Tiers) != 6 {
		t.Fatalf("the curated map has 6 maturity tiers, ids.go carries %d", len(Iteration0Tiers))
	}
}

// TestIDsMatchDataset fails loudly if ids.go drifts from data/atlas.json, which
// is what would happen if someone cut a new iteration without regenerating.
func TestIDsMatchDataset(t *testing.T) {
	raw, err := os.ReadFile(filepath.FromSlash(atlasPath))
	if err != nil {
		t.Skipf("dataset not readable from here (%v); skipping drift check", err)
	}

	var doc struct {
		Agendas []struct {
			ID      string `json:"id"`
			AreaTag string `json:"area_tag"`
		} `json:"agendas"`
		Areas []struct {
			Tag string `json:"tag"`
		} `json:"areas"`
		Problems []struct {
			ID string `json:"id"`
		} `json:"problems"`
		Legend struct {
			TierOrder []string `json:"tier_order"`
		} `json:"legend"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parsing atlas.json: %v", err)
	}

	agendaIDs := make([]string, len(doc.Agendas))
	for i, a := range doc.Agendas {
		agendaIDs[i] = a.ID
	}
	assertSameOrder(t, "agenda ids", agendaIDs, Iteration0AgendaIDs)

	areaTags := make([]string, len(doc.Areas))
	for i, a := range doc.Areas {
		areaTags[i] = a.Tag
	}
	assertSameOrder(t, "area tags", areaTags, Iteration0AreaTags)

	problemIDs := make([]string, len(doc.Problems))
	for i, p := range doc.Problems {
		problemIDs[i] = p.ID
	}
	assertSameOrder(t, "problem ids", problemIDs, Iteration0ProblemIDs)

	// tier_order is the vote vocabulary. A renamed tier silently invalidates every
	// stored vote, so drift here has to fail as loudly as a renamed agenda.
	assertSameOrder(t, "tiers", doc.Legend.TierOrder, Iteration0Tiers)

	for _, a := range doc.Agendas {
		if tag, ok := AreaOfAgenda(a.ID); !ok || tag != a.AreaTag {
			t.Fatalf("area of %s: atlas.json has %q, ids.go has %q (found=%v), run scripts/gen_agenda_ids.py",
				a.ID, a.AreaTag, tag, ok)
		}
	}
}

func assertSameOrder(t *testing.T, label string, fromDataset, generated []string) {
	t.Helper()
	if len(fromDataset) != len(generated) {
		t.Fatalf("%s: atlas.json has %d, ids.go has %d, run scripts/gen_agenda_ids.py",
			label, len(fromDataset), len(generated))
	}
	for i := range fromDataset {
		if fromDataset[i] != generated[i] {
			t.Fatalf("%s[%d]: atlas.json has %q, ids.go has %q, run scripts/gen_agenda_ids.py",
				label, i, fromDataset[i], generated[i])
		}
	}
}
