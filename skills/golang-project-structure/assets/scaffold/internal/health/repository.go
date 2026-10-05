package health

import (
	"context"
	"database/sql"
	"fmt"

	tracing "example.com/app/internal/pkg/observability"
)

// Repository is the only file in the domain that touches the database.
type Repository struct {
	db *sql.DB
}

// NewRepository builds a Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Ping checks database reachability.
func (r *Repository) Ping(ctx context.Context) (err error) {
	ctx, span := tracing.Start(ctx, "health.Repository.Ping")
	defer tracing.End(span, &err)

	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseUnreachable, err)
	}
	return nil
}
