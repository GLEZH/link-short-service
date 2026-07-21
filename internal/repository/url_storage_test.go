package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GLEZH/linkshrtservice/internal/entity"
)

func TestURLStorage_Get(t *testing.T) {
	storage, err := New("")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

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
		if !strings.Contains(err.Error(), "missing") {
			t.Fatalf("Get() error = %v, want id in error", err)
		}
	})
}

func TestURLStorage_Persistence(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "short-url-db.json")
	originalURL := "http://yandex.ru"

	storage, err := New(filePath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	savedURL, err := storage.Save(entity.URL{OriginalURL: originalURL})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := storage.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	var records []record
	if err = json.Unmarshal(data, &records); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("records count = %d, want 1", len(records))
	}

	if records[0].UUID != savedURL.ID {
		t.Errorf("UUID = %q, want %q", records[0].UUID, savedURL.ID)
	}

	if records[0].ShortURL != savedURL.ID {
		t.Errorf("ShortURL = %q, want %q", records[0].ShortURL, savedURL.ID)
	}

	if records[0].OriginalURL != originalURL {
		t.Errorf("OriginalURL = %q, want %q", records[0].OriginalURL, originalURL)
	}

	restored, err := New(filePath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = restored.Close()
	})

	gotURL, err := restored.Get(savedURL.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if gotURL.OriginalURL != originalURL {
		t.Errorf("OriginalURL = %q, want %q", gotURL.OriginalURL, originalURL)
	}
}
