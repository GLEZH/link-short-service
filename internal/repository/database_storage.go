package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/GLEZH/linkshrtservice/internal/entity"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type DatabaseURLStorage struct {
	db *sql.DB
}

func NewDatabaseURLStorage(db *sql.DB) *DatabaseURLStorage {
	return &DatabaseURLStorage{db: db}
}

func (s *DatabaseURLStorage) Save(ctx context.Context, url entity.URL) (entity.URL, error) {
	ctx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()

	var id int64
	err := s.db.QueryRowContext(
		ctx,
		"INSERT INTO shortened_urls (original_url) VALUES ($1) RETURNING id",
		url.OriginalURL,
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			existingURL, getErr := s.getByOriginalURL(ctx, url.OriginalURL)
			if getErr != nil {
				return entity.URL{}, fmt.Errorf("get existing url: %w", getErr)
			}
			return entity.URL{}, entity.NewURLAlreadyExistsError(existingURL)
		}
		return entity.URL{}, fmt.Errorf("save url: %w", err)
	}

	url.ID = strconv.FormatInt(id, 10)
	return url, nil
}

func (s *DatabaseURLStorage) SaveBatch(ctx context.Context, urls []entity.URL) ([]entity.URL, error) {
	ctx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin save batch: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO shortened_urls (original_url)
		VALUES ($1)
		ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
		RETURNING id
	`)
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

func (s *DatabaseURLStorage) getByOriginalURL(ctx context.Context, originalURL string) (entity.URL, error) {
	var id int64
	err := s.db.QueryRowContext(
		ctx,
		"SELECT id FROM shortened_urls WHERE original_url = $1",
		originalURL,
	).Scan(&id)
	if err != nil {
		return entity.URL{}, fmt.Errorf("get url by original url: %w", err)
	}

	return entity.URL{
		ID:          strconv.FormatInt(id, 10),
		OriginalURL: originalURL,
	}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}

func (s *DatabaseURLStorage) Get(ctx context.Context, id string) (entity.URL, error) {
	urlID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return entity.URL{}, entity.NewURLNotFoundError(id)
	}

	ctx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()

	url := entity.URL{ID: id}
	err = s.db.QueryRowContext(
		ctx,
		"SELECT original_url FROM shortened_urls WHERE id = $1",
		urlID,
	).Scan(&url.OriginalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.URL{}, entity.NewURLNotFoundError(id)
		}
		return entity.URL{}, fmt.Errorf("get url: %w", err)
	}

	return url, nil
}
