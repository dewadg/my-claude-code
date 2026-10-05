package health

import "context"

// healthService is the handler's view of the service.
type healthService interface {
	Check(ctx context.Context) Report
}

var _ healthService = (*Service)(nil)
