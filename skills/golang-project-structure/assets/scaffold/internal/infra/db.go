package infra

import (
	"context"
	"database/sql"
	"fmt"

	"example.com/app/internal/pkg/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"
)

// NewDB opens the pool, bounds the initial ping with DB_PING_TIMEOUT and
// registers the close. Constructed lazily, on the first consumer. Swap the
// driver or wrap it in an ORM here; nothing else opens connections.
func NewDB(lc fx.Lifecycle, cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("infra: open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DBPingTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("infra: ping database: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return db.Close()
		},
	})

	return db, nil
}
