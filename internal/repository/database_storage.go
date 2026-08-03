package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/GLEZH/linkshrtservice/internal/entity"
)

type DatabaseURLStorage struct {
	db *sql.DB
}

func NewDatabaseURLStorage(db *sql.DB) *DatabaseURLStorage {
	return &DatabaseURLStorage{db: db}
}

func (s *DatabaseURLStorage) Save(url entity.URL) (entity.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var id int64
	err := s.db.QueryRowContext(
		ctx,
		"INSERT INTO shortened_urls (original_url) VALUES ($1) RETURNING id",
		url.OriginalURL,
	).Scan(&id)
	if err != nil {
		return entity.URL{}, fmt.Errorf("save url: %w", err)
	}

	url.ID = strconv.FormatInt(id, 10)
	return url, nil
}

func (s *DatabaseURLStorage) SaveBatch(urls []entity.URL) ([]entity.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin save batch: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO shortened_urls (original_url) VALUES ($1) RETURNING id")
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("prepare save batch: %w", err)
	}
	defer stmt.Close()

	savedURLs := make([]entity.URL, 0, len(urls))
	for _, url := range urls {
		var id int64
		if err = stmt.QueryRowContext(ctx, url.OriginalURL).Scan(&id); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("save batch url: %w", err)
		}
		url.ID = strconv.FormatInt(id, 10)
		savedURLs = append(savedURLs, url)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit save batch: %w", err)
	}

	return savedURLs, nil
}

func (s *DatabaseURLStorage) Get(id string) (entity.URL, error) {
	urlID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return entity.URL{}, fmt.Errorf("%w: id %s", entity.ErrURLNotFound, id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	url := entity.URL{ID: id}
	err = s.db.QueryRowContext(
		ctx,
		"SELECT original_url FROM shortened_urls WHERE id = $1",
		urlID,
	).Scan(&url.OriginalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.URL{}, fmt.Errorf("%w: id %s", entity.ErrURLNotFound, id)
		}
		return entity.URL{}, fmt.Errorf("get url: %w", err)
	}

	return url, nil
}
