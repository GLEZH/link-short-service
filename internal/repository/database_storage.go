package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

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
		"INSERT INTO shortened_urls (original_url, user_id) VALUES ($1, $2) RETURNING id",
		url.OriginalURL,
		url.UserID,
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
		INSERT INTO shortened_urls (original_url, user_id)
		VALUES ($1, $2)
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
		if err = stmt.QueryRowContext(ctx, url.OriginalURL, url.UserID).Scan(&id); err != nil {
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
	var userID string
	var isDeleted bool
	err := s.db.QueryRowContext(
		ctx,
		"SELECT id, user_id, is_deleted FROM shortened_urls WHERE original_url = $1",
		originalURL,
	).Scan(&id, &userID, &isDeleted)
	if err != nil {
		return entity.URL{}, fmt.Errorf("get url by original url: %w", err)
	}

	return entity.URL{
		ID:          strconv.FormatInt(id, 10),
		OriginalURL: originalURL,
		UserID:      userID,
		IsDeleted:   isDeleted,
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
		"SELECT original_url, is_deleted FROM shortened_urls WHERE id = $1",
		urlID,
	).Scan(&url.OriginalURL, &url.IsDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.URL{}, entity.NewURLNotFoundError(id)
		}
		return entity.URL{}, fmt.Errorf("get url: %w", err)
	}
	if url.IsDeleted {
		return entity.URL{}, entity.NewURLDeletedError(id)
	}

	return url, nil
}

func (s *DatabaseURLStorage) GetByUserID(ctx context.Context, userID string) ([]entity.URL, error) {
	ctx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT id, original_url FROM shortened_urls WHERE user_id = $1 ORDER BY id",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get urls by user id: %w", err)
	}
	defer rows.Close()

	urls := make([]entity.URL, 0)
	for rows.Next() {
		var id int64
		url := entity.URL{UserID: userID}
		if err = rows.Scan(&id, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("scan user url: %w", err)
		}
		url.ID = strconv.FormatInt(id, 10)
		urls = append(urls, url)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("read user urls: %w", err)
	}

	return urls, nil
}

func (s *DatabaseURLStorage) DeleteBatch(ctx context.Context, userID string, ids []string) error {
	ctx, cancel := context.WithTimeout(ctx, storageOperationTimeout)
	defer cancel()

	query, args := buildDeleteBatchQuery(userID, ids)
	if query == "" {
		return nil
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("delete batch urls: %w", err)
	}

	return nil
}

func buildDeleteBatchQuery(userID string, ids []string) (string, []any) {
	args := []any{userID}
	placeholders := make([]string, 0, len(ids))

	for _, id := range ids {
		urlID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			continue
		}
		args = append(args, urlID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
	}
	if len(placeholders) == 0 {
		return "", nil
	}

	query := fmt.Sprintf(
		"UPDATE shortened_urls SET is_deleted = TRUE WHERE user_id = $1 AND id IN (%s)",
		strings.Join(placeholders, ","),
	)

	return query, args
}
