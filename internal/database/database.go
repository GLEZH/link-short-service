package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrEmptyDSN = errors.New("database dsn is empty")

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

func (d *DB) Close() error {
	if d == nil || d.db == nil {
		return nil
	}

	return d.db.Close()
}
