// Command backend serves the feedback API for the AI Safety Agendas x Problems Map.
//
// The curated map itself is static and shipped with the frontend; this service
// only stores and aggregates researcher feedback.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"ai-safety-atlas/backend/internal/api"
	"ai-safety-atlas/backend/internal/config"
	"ai-safety-atlas/backend/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	// Loads backend/.env into the process environment for local `go run .`. It is
	// a no-op wherever the file doesn't exist (Render, Docker, CI), and it never
	// overrides a variable the environment already set, real deployment env vars
	// always win over a stray .env.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		slog.Warn("could not read .env", "error", err)
	}

	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var st api.Storage
	if cfg.Degraded() {
		slog.Warn("DATABASE_URL is not set: running without storage, feedback submission disabled")
	} else {
		db, err := store.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			return err
		}
		defer db.Close()

		migrateCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		if err := db.Migrate(migrateCtx); err != nil {
			return err
		}
		st = db
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.New(cfg, st).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening",
			"port", cfg.Port,
			"allowed_origin", cfg.AllowedOrigin,
			"iteration", cfg.Iteration,
			"turnstile", cfg.TurnstileSecret != "",
			"degraded", cfg.Degraded(),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
