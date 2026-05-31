package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
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
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig: %w", err)
	}

	// SimpleProtocol embeds parameters as text literals, avoiding unnamed prepared
	// statements entirely. DescribeExec (the alternative) reuses the unnamed statement
	// "" per connection; when queries with different parameter counts share a connection
	// the bind count mismatches, producing SQLSTATE 08P01.
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET TIME ZONE 'Asia/Bangkok'")
		return err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
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
