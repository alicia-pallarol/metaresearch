package api

import (
	"net/mail"
	"strings"
)

// submissionRequest is the raw JSON body of a submission.
type submissionRequest struct {
	Name           string          `json:"name"`
	Email          string          `json:"email"`
	Anonymous      bool            `json:"anonymous"`
	ContactConsent bool            `json:"contact_consent"`
	Honeypot       string          `json:"website"` // bots fill this; humans never see it
	Ratings        map[string]*int `json:"ratings"` // tag -> 0..3, null/omitted = blank
}

// validationError maps a field to a human-readable message.
type validationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

const (
	maxNameLen  = 200
	maxEmailLen = 254
)

// validate normalizes and checks the request against tag->id, returning the
// cleaned fields plus the resolved area-id ratings, or a list of errors.
func (r *submissionRequest) validate(tagToID map[string]int) (email, name string, ratings map[int]int, errs []validationError) {
	email = strings.TrimSpace(strings.ToLower(r.Email))
	name = strings.TrimSpace(r.Name)

	if email == "" {
		errs = append(errs, validationError{"email", "Email is required."})
	} else if len(email) > maxEmailLen {
		errs = append(errs, validationError{"email", "Email is too long."})
	} else if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		errs = append(errs, validationError{"email", "Please enter a valid email address."})
	}

	if len(name) > maxNameLen {
		errs = append(errs, validationError{"name", "Name is too long."})
	}
	// A name is required only when the submitter is not anonymous.
	if !r.Anonymous && name == "" {
		errs = append(errs, validationError{"name", "Name is required unless you submit anonymously."})
	}

	ratings = make(map[int]int)
	for tag, val := range r.Ratings {
		if val == nil {
			continue // blank is allowed
		}
		id, ok := tagToID[tag]
		if !ok {
			errs = append(errs, validationError{"ratings", "Unknown research area: " + tag})
			continue
		}
		if *val < 0 || *val > 3 {
			errs = append(errs, validationError{"ratings", "Familiarity must be between 0 and 3."})
			continue
		}
		ratings[id] = *val
	}

	return email, name, ratings, errs
}
