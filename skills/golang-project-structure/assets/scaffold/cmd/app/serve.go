package main

import (
	"log/slog"

	"example.com/app/internal/health"
	"example.com/app/internal/httpserver"
	"example.com/app/internal/infra"
	"example.com/app/internal/pkg/config"
	"example.com/app/internal/pkg/logging"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			l := slog.With(logging.KeySource, "main.serve")

			cfg, err := config.Parse()
			if err != nil {
				l.Error("config parse failed", logging.KeyError, err)
				return err
			}
			if err := config.AssertRequiredEnv(cfg); err != nil {
				l.Error("config validation failed", logging.KeyError, err)
				return err
			}

			// The only module registration point. Order is for reading; fx
			// resolves by type.
			fx.New(
				fx.Supply(cfg),
				fx.StopTimeout(cfg.ShutdownTimeout),
				infra.Module,
				health.Module,
				httpserver.Module,
			).Run()
			return nil
		},
	}
}
