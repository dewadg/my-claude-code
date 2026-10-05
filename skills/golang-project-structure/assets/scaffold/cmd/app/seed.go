package main

import (
	"context"
	"database/sql"
	"log/slog"

	"example.com/app/internal/pkg/logging"
	"example.com/app/internal/seeders"
	"github.com/spf13/cobra"
)

func newSeedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Run idempotent data seeders (run migrate first)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withDB(cmd.Context(), func(ctx context.Context, db *sql.DB) error {
				if err := seeders.Run(ctx, db); err != nil {
					return err
				}
				slog.With(logging.KeySource, "main.seed").InfoContext(ctx, "seed completed")
				return nil
			})
		},
	}
}
