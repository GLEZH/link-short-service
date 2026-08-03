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
