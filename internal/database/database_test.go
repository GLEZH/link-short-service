package database

import (
	"context"
	"errors"
	"testing"
)

func TestNew_EmptyDSN(t *testing.T) {
	db, err := New("")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if db != nil {
		t.Fatalf("New() db = %+v, want nil", db)
	}
}

func TestNew_WithDSN(t *testing.T) {
	db, err := New("postgres://shortener:shortener@localhost:5432/shortener?sslmode=disable")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("New() db = nil, want db")
	}

	if sqlDB := db.SQLDB(); sqlDB == nil {
		t.Fatal("SQLDB() = nil, want db")
	}
}

func TestDB_NilConnection(t *testing.T) {
	var db *DB

	if err := db.Ping(context.Background()); !errors.Is(err, ErrEmptyDSN) {
		t.Fatalf("Ping() error = %v, want %v", err, ErrEmptyDSN)
	}

	if err := db.Migrate(); !errors.Is(err, ErrEmptyDSN) {
		t.Fatalf("Migrate() error = %v, want %v", err, ErrEmptyDSN)
	}

	if sqlDB := db.SQLDB(); sqlDB != nil {
		t.Fatalf("SQLDB() = %+v, want nil", sqlDB)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
