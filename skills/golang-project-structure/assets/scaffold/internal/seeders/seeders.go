package seeders

import (
	"context"
	"database/sql"
	"log/slog"

	"example.com/app/internal/pkg/buildutil"
	"example.com/app/internal/pkg/logging"
)

// Seeder is one idempotent seeding step.
type Seeder struct {
	Name string
	Run  func(ctx context.Context, db *sql.DB) error
}

// All lists seeders in execution order. Every step must be safe to re-run.
func All() []Seeder {
	return []Seeder{
		{Name: "app_meta", Run: seedAppMeta},
	}
}

// Run executes every seeder in order and stops at the first error.
func Run(ctx context.Context, db *sql.DB) error {
	l := slog.With(logging.KeySource, "seeders.Run")

	for _, s := range All() {
		l.InfoContext(ctx, "seeder started", "seeder", s.Name)
		if err := s.Run(ctx, db); err != nil {
			l.ErrorContext(ctx, "seeder failed", "seeder", s.Name, logging.KeyError, err)
			return err
		}
	}
	return nil
}

func seedAppMeta(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO app_meta (key, value) VALUES ('seeded_by', $1) ON CONFLICT (key) DO NOTHING`,
		buildutil.Name+"@"+buildutil.Version,
	)
	return err
}
