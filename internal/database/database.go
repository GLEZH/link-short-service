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

var ErrEmptyDSN = errors.New("database dsn is empty")

//go:embed migrations/*.sql
var migrations embed.FS

type DB struct {
	db *sql.DB
}

func New(dsn string) (*DB, error) {
	if dsn == "" {
		return &DB{}, nil
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

func (d *DB) Ping(ctx context.Context) error {
	if d == nil || d.db == nil {
		return ErrEmptyDSN
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	return d.db.PingContext(ctx)
}

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

func (d *DB) SQLDB() *sql.DB {
	if d == nil {
		return nil
	}

	return d.db
}

func (d *DB) Close() error {
	if d == nil || d.db == nil {
		return nil
	}

	return d.db.Close()
}
