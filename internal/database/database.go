package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// ErrEmptyDSN reports that no database connection is configured
var ErrEmptyDSN = errors.New("database dsn is empty")

//go:embed migrations/*.sql
var migrations embed.FS

// DB wraps a PostgreSQL connection pool
type DB struct {
	db *sql.DB
}

// New opens a PostgreSQL connection pool
func New(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, nil
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

// Ping checks the database connection
func (d *DB) Ping(ctx context.Context) error {
	if d == nil || d.db == nil {
		return ErrEmptyDSN
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	return d.db.PingContext(ctx)
}

// Migrate applies pending database migrations
func (d *DB) Migrate() error {
	if d == nil || d.db == nil {
		return ErrEmptyDSN
	}

	goose.SetBaseFS(migrations)
	defer goose.SetBaseFS(nil)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(d.db, "migrations")
}

// SQLDB returns the underlying connection pool
func (d *DB) SQLDB() *sql.DB {
	if d == nil {
		return nil
	}

	return d.db
}

// Close closes the database connection pool
func (d *DB) Close() error {
	if d == nil || d.db == nil {
		return nil
	}

	return d.db.Close()
}
