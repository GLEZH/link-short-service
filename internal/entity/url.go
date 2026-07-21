package entity

import "errors"

var (
	ErrInvalidURL  = errors.New("invalid url")
	ErrURLNotFound = errors.New("url not found")
)

type URL struct {
	ID          string
	OriginalURL string
}
