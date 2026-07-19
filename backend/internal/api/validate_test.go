package api

import "testing"

func ptr(i int) *int { return &i }

func TestValidate(t *testing.T) {
	tagToID := map[string]int{"EA": 1, "IN": 2}

	tests := []struct {
		name       string
		req        submissionRequest
		wantErrs   int
		wantEmail  string
		wantName   string
		wantRating map[int]int
	}{
		{
			name: "valid non-anonymous",
			req: submissionRequest{
				Name: "  Ada Lovelace ", Email: "Ada@Example.COM ",
				Ratings: map[string]*int{"EA": ptr(2), "IN": nil},
			},
			wantErrs: 0, wantEmail: "ada@example.com", wantName: "Ada Lovelace",
			wantRating: map[int]int{1: 2},
		},
		{
			name:     "valid anonymous without name",
			req:      submissionRequest{Email: "x@y.io", Anonymous: true},
			wantErrs: 0, wantEmail: "x@y.io", wantName: "",
			wantRating: map[int]int{},
		},
		{
			name:     "missing email",
			req:      submissionRequest{Anonymous: true},
			wantErrs: 1,
		},
		{
			name:     "invalid email",
			req:      submissionRequest{Email: "not-an-email", Anonymous: true},
			wantErrs: 1,
		},
		{
			name:     "non-anonymous requires name",
			req:      submissionRequest{Email: "a@b.io", Anonymous: false},
			wantErrs: 1,
		},
		{
			name:     "rating out of range",
			req:      submissionRequest{Email: "a@b.io", Anonymous: true, Ratings: map[string]*int{"EA": ptr(5)}},
			wantErrs: 1,
		},
		{
			name:     "unknown tag",
			req:      submissionRequest{Email: "a@b.io", Anonymous: true, Ratings: map[string]*int{"ZZ": ptr(1)}},
			wantErrs: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			email, name, ratings, errs := tc.req.validate(tagToID)
			if len(errs) != tc.wantErrs {
				t.Fatalf("got %d errors %v, want %d", len(errs), errs, tc.wantErrs)
			}
			if tc.wantErrs > 0 {
				return
			}
			if email != tc.wantEmail {
				t.Errorf("email = %q, want %q", email, tc.wantEmail)
			}
			if name != tc.wantName {
				t.Errorf("name = %q, want %q", name, tc.wantName)
			}
			if len(ratings) != len(tc.wantRating) {
				t.Errorf("ratings = %v, want %v", ratings, tc.wantRating)
			}
			for k, v := range tc.wantRating {
				if ratings[k] != v {
					t.Errorf("ratings[%d] = %d, want %d", k, ratings[k], v)
				}
			}
		})
	}
}
