package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"example.com/app/build/migrations"
	"example.com/app/internal/pkg/logging"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

// migrationsDir is where `migrate create` writes, relative to the module root.
const migrationsDir = "build/migrations"

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply all pending migrations (idempotent)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withProvider(cmd.Context(), func(ctx context.Context, p *goose.Provider) error {
				l := slog.With(logging.KeySource, "main.migrate.up")
				results, err := p.Up(ctx)
				for _, r := range results {
					l.InfoContext(ctx, "migration applied", "version", r.Source.Version, logging.KeyDurationMS, r.Duration.Milliseconds())
				}
				return err
			})
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "down",
		Short: "Roll back the most recent migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withProvider(cmd.Context(), func(ctx context.Context, p *goose.Provider) error {
				l := slog.With(logging.KeySource, "main.migrate.down")
				r, err := p.Down(ctx)
				if r != nil && r.Source != nil {
					l.InfoContext(ctx, "migration rolled back", "version", r.Source.Version)
				}
				return err
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Print applied and pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withProvider(cmd.Context(), func(ctx context.Context, p *goose.Provider) error {
				statuses, err := p.Status(ctx)
				if err != nil {
					return err
				}
				for _, s := range statuses {
					fmt.Printf("%-8s %05d %s\n", s.State, s.Source.Version, s.Source.Path)
				}
				return nil
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "create <name>",
		Short: "Create the next sequential SQL migration in " + migrationsDir,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			goose.SetSequential(true)
			return goose.Create(nil, migrationsDir, args[0], "sql")
		},
	})

	return cmd
}

// withProvider runs fn against a goose provider over the embedded
// migrations. migrate runs as a one-off process before deploy and never
// starts the fx graph.
func withProvider(ctx context.Context, fn func(context.Context, *goose.Provider) error) error {
	l := slog.With(logging.KeySource, "main.migrate")

	return withDB(ctx, func(ctx context.Context, db *sql.DB) error {
		p, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
		if err != nil {
			l.ErrorContext(ctx, "goose provider build failed", logging.KeyError, err)
			return err
		}
		if err := fn(ctx, p); err != nil {
			l.ErrorContext(ctx, "migrate failed", logging.KeyError, err)
			return err
		}
		return nil
	})
}
