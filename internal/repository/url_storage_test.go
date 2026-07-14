package repository

import (
	"errors"
	"testing"

	"github.com/GLEZH/linkshrtservice/internal/entity"
)

func TestURLStorage_Get(t *testing.T) {
	storage := NewURLStorage()

	t.Run("existing url", func(t *testing.T) {
		savedURL, err := storage.Save(entity.URL{OriginalURL: "http://example.com"})
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		gotURL, err := storage.Get(savedURL.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if gotURL.OriginalURL != savedURL.OriginalURL {
			t.Errorf("OriginalURL = %q, want %q", gotURL.OriginalURL, savedURL.OriginalURL)
		}
	})

	t.Run("missing url", func(t *testing.T) {
		_, err := storage.Get("missing")
		if !errors.Is(err, entity.ErrURLNotFound) {
			t.Fatalf("Get() error = %v, want %v", err, entity.ErrURLNotFound)
		}
	})
}
