package entity

import (
	"errors"
	"testing"
)

func TestDomainErrors(t *testing.T) {
	t.Run("invalid url", func(t *testing.T) {
		err := NewInvalidURLError()

		if !errors.Is(err, ErrInvalidURL) {
			t.Fatalf("error = %v, want %v", err, ErrInvalidURL)
		}

		if err.Code() != "url.invalid" {
			t.Errorf("Code() = %q, want %q", err.Code(), "url.invalid")
		}
	})

	t.Run("url not found", func(t *testing.T) {
		err := NewURLNotFoundError("missing")

		if !errors.Is(err, ErrURLNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrURLNotFound)
		}

		if err.ID != "missing" {
			t.Errorf("ID = %q, want %q", err.ID, "missing")
		}
	})

	t.Run("url already exists", func(t *testing.T) {
		url := URL{ID: "1", OriginalURL: "http://example.com"}
		err := NewURLAlreadyExistsError(url)

		if !errors.Is(err, ErrURLAlreadyExists) {
			t.Fatalf("error = %v, want %v", err, ErrURLAlreadyExists)
		}

		var alreadyExists *URLAlreadyExistsError
		if !errors.As(err, &alreadyExists) {
			t.Fatal("errors.As() = false, want true")
		}

		if alreadyExists.URL != url {
			t.Errorf("URL = %+v, want %+v", alreadyExists.URL, url)
		}
	})
}
