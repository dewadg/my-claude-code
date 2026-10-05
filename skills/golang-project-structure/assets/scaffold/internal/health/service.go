package health

import (
	"context"
	"log/slog"

	"example.com/app/internal/pkg/buildutil"
	"example.com/app/internal/pkg/logging"
	tracing "example.com/app/internal/pkg/observability"
)

// Service runs dependency checks.
type Service struct {
	store healthStore
}

// NewService builds a Service.
func NewService(store healthStore) *Service {
	return &Service{store: store}
}

// Check never fails; a failed dependency degrades the report.
func (s *Service) Check(ctx context.Context) Report {
	ctx, span := tracing.Start(ctx, "health.Service.Check")
	defer span.End()

	l := slog.With(logging.KeySource, "health.Service.Check")

	report := Report{Status: StatusOK, Version: buildutil.Version, Checks: map[string]string{"database": StatusOK}}
	if err := s.store.Ping(ctx); err != nil {
		l.WarnContext(ctx, "database check failed", logging.KeyError, err)
		report.Status = StatusDegraded
		report.Checks["database"] = StatusDegraded
	}
	return report
}
