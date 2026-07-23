package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// turnstileVerifier checks a Turnstile token with Cloudflare. It is only
// constructed when TURNSTILE_SECRET is set; the MVP runs fine without it.
type turnstileVerifier struct {
	secret   string
	endpoint string
	client   *http.Client
}

func newTurnstileVerifier(secret string) *turnstileVerifier {
	return &turnstileVerifier{
		secret:   secret,
		endpoint: turnstileVerifyURL,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

// Verify returns nil when Cloudflare accepts the token.
func (v *turnstileVerifier) Verify(ctx context.Context, token, remoteIP string) error {
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("missing turnstile token")
	}
	form := url.Values{}
	form.Set("secret", v.secret)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("building turnstile request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("calling turnstile: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("decoding turnstile response: %w", err)
	}
	if !body.Success {
		return fmt.Errorf("turnstile rejected the token: %s", strings.Join(body.ErrorCodes, ","))
	}
	return nil
}
