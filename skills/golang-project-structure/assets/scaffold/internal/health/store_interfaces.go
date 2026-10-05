package health

import "context"

// healthStore is the service's view of the repository.
type healthStore interface {
	Ping(ctx context.Context) error
}

var _ healthStore = (*Repository)(nil)
