package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// DB holds both the pgxpool (for connection management) and a
// database/sql wrapper needed by go-jet/jet's query runner.
type DB struct {
	Pool *pgxpool.Pool
	db   *sql.DB
}

func New(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return &DB{Pool: pool, db: stdlib.OpenDBFromPool(pool)}, nil
}

// SQL returns the database/sql handle used by jet query statements.
func (d *DB) SQL() *sql.DB { return d.db }

// Close shuts down both the sql.DB wrapper and the underlying pool.
func (d *DB) Close() {
	_ = d.db.Close()
	d.Pool.Close()
}
