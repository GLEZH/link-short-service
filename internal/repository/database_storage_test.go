package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GLEZH/linkshrtservice/internal/entity"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func newMockDatabaseStorage(t *testing.T) (*DatabaseURLStorage, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}

	return NewDatabaseURLStorage(db), mock, func() {
		_ = db.Close()
	}
}

func TestDatabaseURLStorage_Save(t *testing.T) {
	ctx := context.Background()

	t.Run("new url", func(t *testing.T) {
		storage, mock, closeDB := newMockDatabaseStorage(t)
		defer closeDB()

		mock.ExpectQuery("INSERT INTO shortened_urls").
			WithArgs("http://example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		savedURL, err := storage.Save(ctx, entity.URL{OriginalURL: "http://example.com"})
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		if savedURL.ID != "1" {
			t.Errorf("ID = %q, want %q", savedURL.ID, "1")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations error = %v", err)
		}
	})

	t.Run("existing url", func(t *testing.T) {
		storage, mock, closeDB := newMockDatabaseStorage(t)
		defer closeDB()

		mock.ExpectQuery("INSERT INTO shortened_urls").
			WithArgs("http://example.com").
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
		mock.ExpectQuery("SELECT id FROM shortened_urls").
			WithArgs("http://example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

		_, err := storage.Save(ctx, entity.URL{OriginalURL: "http://example.com"})
		if !errors.Is(err, entity.ErrURLAlreadyExists) {
			t.Fatalf("Save() error = %v, want %v", err, entity.ErrURLAlreadyExists)
		}

		var alreadyExists *entity.URLAlreadyExistsError
		if !errors.As(err, &alreadyExists) {
			t.Fatal("errors.As() = false, want true")
		}

		if alreadyExists.URL.ID != "7" {
			t.Errorf("existing ID = %q, want %q", alreadyExists.URL.ID, "7")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations error = %v", err)
		}
	})

	t.Run("insert error", func(t *testing.T) {
		storage, mock, closeDB := newMockDatabaseStorage(t)
		defer closeDB()

		mock.ExpectQuery("INSERT INTO shortened_urls").
			WithArgs("http://example.com").
			WillReturnError(errors.New("insert failed"))

		_, err := storage.Save(ctx, entity.URL{OriginalURL: "http://example.com"})
		if err == nil {
			t.Fatal("Save() error = nil, want error")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations error = %v", err)
		}
	})

	t.Run("existing url get error", func(t *testing.T) {
		storage, mock, closeDB := newMockDatabaseStorage(t)
		defer closeDB()

		mock.ExpectQuery("INSERT INTO shortened_urls").
			WithArgs("http://example.com").
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
		mock.ExpectQuery("SELECT id FROM shortened_urls").
			WithArgs("http://example.com").
			WillReturnError(sql.ErrNoRows)

		_, err := storage.Save(ctx, entity.URL{OriginalURL: "http://example.com"})
		if err == nil {
			t.Fatal("Save() error = nil, want error")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations error = %v", err)
		}
	})
}

func TestDatabaseURLStorage_SaveBatch(t *testing.T) {
	ctx := context.Background()
	storage, mock, closeDB := newMockDatabaseStorage(t)
	defer closeDB()

	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO shortened_urls")
	mock.ExpectQuery("INSERT INTO shortened_urls").
		WithArgs("http://example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("INSERT INTO shortened_urls").
		WithArgs("http://practicum.yandex.ru").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()

	savedURLs, err := storage.SaveBatch(ctx, []entity.URL{
		{OriginalURL: "http://example.com"},
		{OriginalURL: "http://practicum.yandex.ru"},
	})
	if err != nil {
		t.Fatalf("SaveBatch() error = %v", err)
	}

	if len(savedURLs) != 2 {
		t.Fatalf("saved urls count = %d, want 2", len(savedURLs))
	}

	if savedURLs[0].ID != "1" || savedURLs[1].ID != "2" {
		t.Fatalf("saved ids = %q, %q; want 1, 2", savedURLs[0].ID, savedURLs[1].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations error = %v", err)
	}
}

func TestDatabaseURLStorage_SaveBatchErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("begin error", func(t *testing.T) {
		storage, mock, closeDB := newMockDatabaseStorage(t)
		defer closeDB()

		mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

		_, err := storage.SaveBatch(ctx, []entity.URL{{OriginalURL: "http://example.com"}})
		if err == nil {
			t.Fatal("SaveBatch() error = nil, want error")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations error = %v", err)
		}
	})

	t.Run("prepare error", func(t *testing.T) {
		storage, mock, closeDB := newMockDatabaseStorage(t)
		defer closeDB()

		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT INTO shortened_urls").WillReturnError(errors.New("prepare failed"))
		mock.ExpectRollback()

		_, err := storage.SaveBatch(ctx, []entity.URL{{OriginalURL: "http://example.com"}})
		if err == nil {
			t.Fatal("SaveBatch() error = nil, want error")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations error = %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		storage, mock, closeDB := newMockDatabaseStorage(t)
		defer closeDB()

		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT INTO shortened_urls")
		mock.ExpectQuery("INSERT INTO shortened_urls").
			WithArgs("http://example.com").
			WillReturnError(errors.New("query failed"))
		mock.ExpectRollback()

		_, err := storage.SaveBatch(ctx, []entity.URL{{OriginalURL: "http://example.com"}})
		if err == nil {
			t.Fatal("SaveBatch() error = nil, want error")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations error = %v", err)
		}
	})

	t.Run("commit error", func(t *testing.T) {
		storage, mock, closeDB := newMockDatabaseStorage(t)
		defer closeDB()

		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT INTO shortened_urls")
		mock.ExpectQuery("INSERT INTO shortened_urls").
			WithArgs("http://example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

		_, err := storage.SaveBatch(ctx, []entity.URL{{OriginalURL: "http://example.com"}})
		if err == nil {
			t.Fatal("SaveBatch() error = nil, want error")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations error = %v", err)
		}
	})
}

func TestDatabaseURLStorage_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("bad id", func(t *testing.T) {
		storage, _, closeDB := newMockDatabaseStorage(t)
		defer closeDB()

		_, err := storage.Get(ctx, "bad")
		if !errors.Is(err, entity.ErrURLNotFound) {
			t.Fatalf("Get() error = %v, want %v", err, entity.ErrURLNotFound)
		}
	})

	t.Run("existing url", func(t *testing.T) {
		storage, mock, closeDB := newMockDatabaseStorage(t)
		defer closeDB()

		mock.ExpectQuery("SELECT original_url FROM shortened_urls").
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"original_url"}).AddRow("http://example.com"))

		gotURL, err := storage.Get(ctx, "1")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if gotURL.OriginalURL != "http://example.com" {
			t.Errorf("OriginalURL = %q, want %q", gotURL.OriginalURL, "http://example.com")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations error = %v", err)
		}
	})

	t.Run("missing url", func(t *testing.T) {
		storage, mock, closeDB := newMockDatabaseStorage(t)
		defer closeDB()

		mock.ExpectQuery("SELECT original_url FROM shortened_urls").
			WithArgs(int64(1)).
			WillReturnError(sql.ErrNoRows)

		_, err := storage.Get(ctx, "1")
		if !errors.Is(err, entity.ErrURLNotFound) {
			t.Fatalf("Get() error = %v, want %v", err, entity.ErrURLNotFound)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations error = %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		storage, mock, closeDB := newMockDatabaseStorage(t)
		defer closeDB()

		mock.ExpectQuery("SELECT original_url FROM shortened_urls").
			WithArgs(int64(1)).
			WillReturnError(errors.New("query failed"))

		_, err := storage.Get(ctx, "1")
		if err == nil {
			t.Fatal("Get() error = nil, want error")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations error = %v", err)
		}
	})
}
