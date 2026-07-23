package api

import (
	"context"
	"net/http"
	"time"
)

// contextWithTimeout bounds a handler's downstream work by the smaller of the
// request's own lifetime and d.
func contextWithTimeout(r *http.Request, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), d)
}
