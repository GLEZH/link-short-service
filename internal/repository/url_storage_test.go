package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GLEZH/linkshrtservice/internal/entity"
)

func TestURLStorage_Get(t *testing.T) {
	ctx := context.Background()
	storage, err := New("")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	t.Run("existing url", func(t *testing.T) {
		savedURL, err := storage.Save(ctx, entity.URL{OriginalURL: "http://example.com"})
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		gotURL, err := storage.Get(ctx, savedURL.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if gotURL.OriginalURL != savedURL.OriginalURL {
			t.Errorf("OriginalURL = %q, want %q", gotURL.OriginalURL, savedURL.OriginalURL)
		}
	})

	t.Run("missing url", func(t *testing.T) {
		_, err := storage.Get(ctx, "missing")
		if !errors.Is(err, entity.ErrURLNotFound) {
			t.Fatalf("Get() error = %v, want %v", err, entity.ErrURLNotFound)
		}
		if !strings.Contains(err.Error(), "missing") {
			t.Fatalf("Get() error = %v, want id in error", err)
		}
	})
}

func TestURLStorage_Persistence(t *testing.T) {
	ctx := context.Background()
	filePath := filepath.Join(t.TempDir(), "short-url-db.json")
	originalURL := "http://yandex.ru"

	storage, err := New(filePath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	savedURL, err := storage.Save(ctx, entity.URL{OriginalURL: originalURL, UserID: "user-id"})
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

	if records[0].UserID != "user-id" {
		t.Errorf("UserID = %q, want %q", records[0].UserID, "user-id")
	}

	restored, err := New(filePath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = restored.Close()
	})

	gotURL, err := restored.Get(ctx, savedURL.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if gotURL.OriginalURL != originalURL {
		t.Errorf("OriginalURL = %q, want %q", gotURL.OriginalURL, originalURL)
	}
}

func TestURLStorage_GetByUserID(t *testing.T) {
	ctx := context.Background()
	storage, err := New("")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = storage.Save(ctx, entity.URL{OriginalURL: "http://first.example.com", UserID: "first-user"})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	_, err = storage.Save(ctx, entity.URL{OriginalURL: "http://second.example.com", UserID: "second-user"})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	urls, err := storage.GetByUserID(ctx, "first-user")
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}

	if len(urls) != 1 {
		t.Fatalf("urls count = %d, want 1", len(urls))
	}
	if urls[0].OriginalURL != "http://first.example.com" {
		t.Errorf("OriginalURL = %q, want %q", urls[0].OriginalURL, "http://first.example.com")
	}
}

func TestURLStorage_BatchPersistence(t *testing.T) {
	ctx := context.Background()
	filePath := filepath.Join(t.TempDir(), "short-url-db.json")

	storage, err := New(filePath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	savedURLs, err := storage.SaveBatch(ctx, []entity.URL{
		{OriginalURL: "http://yandex.ru", UserID: "user-id"},
		{OriginalURL: "http://practicum.yandex.ru", UserID: "user-id"},
	})
	if err != nil {
		t.Fatalf("SaveBatch() error = %v", err)
	}
	if len(savedURLs) != 2 {
		t.Fatalf("saved urls count = %d, want 2", len(savedURLs))
	}

	restored, err := New(filePath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	gotURL, err := restored.Get(ctx, savedURLs[1].ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if gotURL.OriginalURL != "http://practicum.yandex.ru" {
		t.Errorf("OriginalURL = %q, want %q", gotURL.OriginalURL, "http://practicum.yandex.ru")
	}
}
