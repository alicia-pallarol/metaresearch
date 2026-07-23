package api

import (
	"io"
	"log/slog"
	"os"
	"testing"
)

// TestMain silences request logging so test output stays readable.
func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}
