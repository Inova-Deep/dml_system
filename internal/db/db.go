package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB encapsulates the pgxpool connection
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new database connection pool
func New(ctx context.Context, dsn string) (*DB, error) {
	// Set reasonable timeout for connection attempt
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Default pool config tuning can be done here.
	// E.g., Max number of specific connections
	config.MaxConns = 25
	config.MinConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	// Verify the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	log.Println("Database connection established successfully")
	return &DB{Pool: pool}, nil
}

// Close gracefully shuts down the connection pool
func (db *DB) Close() {
	db.Pool.Close()
	log.Println("Database connection pool closed")
}
