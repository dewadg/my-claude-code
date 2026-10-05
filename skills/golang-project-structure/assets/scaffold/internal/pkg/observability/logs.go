package observability

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/app/internal/pkg/buildutil"
	"example.com/app/internal/pkg/config"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// SetupLogs wraps base with correlation injection and, when logs export is
// on, fans out to the OTel slog bridge behind a level gate pinned to
// cfg.LogLevel. The caller installs the result with slog.SetDefault and must
// call shutdown before exit.
func SetupLogs(ctx context.Context, cfg config.Config, base *slog.Logger) (*slog.Logger, func(context.Context) error, error) {
	if !cfg.OTelLogsEnabled {
		return slog.New(newCorrelationHandler(base.Handler())), func(context.Context) error { return nil }, nil
	}

	setServiceIdentityEnv(cfg.Env)

	exporter, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("observability: create OTLP log exporter: %w", err)
	}
	lp := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)))

	fanOut := slogmulti.Fanout(
		base.Handler(),
		otelslog.NewHandler(buildutil.Name, otelslog.WithLoggerProvider(lp)),
	)

	return slog.New(newCorrelationHandler(newLevelHandler(fanOut, cfg.LogLevel))), lp.Shutdown, nil
}
