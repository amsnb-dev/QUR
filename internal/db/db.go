package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool and exposes tenant-aware helpers.
type Pool struct{ *pgxpool.Pool }

// Connect creates and verifies a pgxpool connection.
func Connect(ctx context.Context, dsn string) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("db parse config: %w", err)
	}
	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	// Initialize BOTH GUC variables at connection level so PostgreSQL
	// recognises them before SET LOCAL is used inside transactions.
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, `
			SELECT set_config('app.school_id',      '', false),
			       set_config('app.is_super_admin', '0', false)
		`)
		return err
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("db create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return &Pool{pool}, nil
}

// TxWithTenant begins a transaction and sets two LOCAL config variables:
//   - app.school_id:      UUID string of the tenant (empty for super_admin)
//   - app.is_super_admin: '1' for super_admin, '0' for everyone else
//
// LOCAL means the values are destroyed when the transaction ends,
// so they can never leak to the next request via pgxpool connection reuse.
func (p *Pool) TxWithTenant(ctx context.Context, schoolID string, isSuperAdmin bool) (pgx.Tx, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	superVal := "0"
	if isSuperAdmin {
		superVal = "1"
	}

	if _, err = tx.Exec(ctx, `
		SELECT set_config('app.school_id',      $1, true),
		       set_config('app.is_super_admin', $2, true)
	`, schoolID, superVal); err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("set tenant: %w", err)
	}

	return tx, nil
}
