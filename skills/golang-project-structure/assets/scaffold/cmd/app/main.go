package main

import (
	"log/slog"
	"os"
	_ "time/tzdata"

	"example.com/app/internal/pkg/buildutil"
	"example.com/app/internal/pkg/config"
	"github.com/spf13/cobra"
)

func main() {
	// Bootstrap logger until infra.Module installs the configured one.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	_ = config.LoadDotenv()

	// Bare binary means serve, so `./app` and `go run ./cmd/app` boot the server.
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "serve")
	}

	rootCmd := &cobra.Command{
		Use:          buildutil.Name,
		Short:        buildutil.Name + " server",
		Version:      buildutil.Version,
		SilenceUsage: true,
	}

	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newMigrateCmd())
	rootCmd.AddCommand(newSeedCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
