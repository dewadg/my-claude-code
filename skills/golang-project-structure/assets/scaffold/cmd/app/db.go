package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"example.com/app/internal/pkg/config"
	"example.com/app/internal/pkg/logging"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// withDB opens a short-lived connection for one-off commands (migrate, seed)
// that run outside the fx graph.
func withDB(ctx context.Context, fn func(context.Context, *sql.DB) error) error {
	l := slog.With(logging.KeySource, "main.withDB")

	cfg, err := config.Parse()
	if err != nil {
		l.ErrorContext(ctx, "config parse failed", logging.KeyError, err)
		return err
	}
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		l.ErrorContext(ctx, "database unreachable", logging.KeyError, err)
		return err
	}
	return fn(ctx, db)
}
