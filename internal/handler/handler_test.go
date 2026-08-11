package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GLEZH/linkshrtservice/internal/auth"
	"github.com/GLEZH/linkshrtservice/internal/entity"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type brokenStorage struct{}

func (s brokenStorage) Save(ctx context.Context, url entity.URL) (entity.URL, error) {
	return entity.URL{}, errors.New("save failed")
}

func (s brokenStorage) SaveBatch(ctx context.Context, urls []entity.URL) ([]entity.URL, error) {
	return nil, errors.New("save batch failed")
}

func (s brokenStorage) Get(ctx context.Context, id string) (entity.URL, error) {
	return entity.URL{}, errors.New("get failed")
}

func (s brokenStorage) GetByUserID(ctx context.Context, userID string) ([]entity.URL, error) {
	return nil, errors.New("get by user failed")
}

func TestShortenURL_InternalError(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	handlers := New("http://localhost:8080", brokenStorage{}, zap.New(core).Sugar(), nil)

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com"))
	request = request.WithContext(auth.WithUserID(request.Context(), "user-id"))
	recorder := httptest.NewRecorder()

	handlers.ShortenURL(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}

	if recorder.Body.String() != "" {
		t.Errorf("body = %q, want empty", recorder.Body.String())
	}

	if logs.Len() != 1 {
		t.Fatalf("logs count = %d, want 1", logs.Len())
	}
}
