package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GLEZH/linkshrtservice/internal/audit"
	"github.com/GLEZH/linkshrtservice/internal/auth"
	"github.com/GLEZH/linkshrtservice/internal/entity"
	"github.com/GLEZH/linkshrtservice/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type brokenStorage struct{}

type recordingAuditor struct {
	events []audit.Event
}

func (a *recordingAuditor) Notify(_ context.Context, event audit.Event) {
	a.events = append(a.events, event)
}

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

func (s brokenStorage) DeleteBatch(ctx context.Context, userID string, ids []string) error {
	return errors.New("delete batch failed")
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

func TestHandler_Audit(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		contentType string
		body        string
		wantURL     string
	}{
		{
			name:    "plain shorten",
			path:    "/",
			body:    "https://example.com/plain",
			wantURL: "https://example.com/plain",
		},
		{
			name:        "json shorten",
			path:        "/api/shorten",
			contentType: "application/json",
			body:        `{"url":"https://example.com/json"}`,
			wantURL:     "https://example.com/json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			auditor := &recordingAuditor{}
			handlers := New("http://localhost:8080", repository.NewURLStorage(), zap.NewNop().Sugar(), nil, auditor)
			router := chi.NewRouter()
			router.Post("/", handlers.ShortenURL)
			router.Post("/api/shorten", handlers.ShortenURLJSON)

			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", test.contentType)
			request = request.WithContext(auth.WithUserID(request.Context(), "user-id"))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusCreated {
				t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusCreated)
			}
			assertAuditEvent(t, auditor.events, audit.ActionShorten, "user-id", test.wantURL)
		})
	}

	t.Run("follow", func(t *testing.T) {
		auditor := &recordingAuditor{}
		storage := repository.NewURLStorage()
		savedURL, err := storage.Save(context.Background(), entity.URL{
			OriginalURL: "https://example.com/follow",
			UserID:      "owner-id",
		})
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		handlers := New("http://localhost:8080", storage, zap.NewNop().Sugar(), nil, auditor)
		router := chi.NewRouter()
		router.Get("/{id}", handlers.GetURL)
		request := httptest.NewRequest(http.MethodGet, "/"+savedURL.ID, nil)
		request = request.WithContext(auth.WithUserID(request.Context(), "visitor-id"))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusTemporaryRedirect {
			t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusTemporaryRedirect)
		}
		assertAuditEvent(t, auditor.events, audit.ActionFollow, "visitor-id", "https://example.com/follow")
	})
}

func assertAuditEvent(t *testing.T, events []audit.Event, action audit.Action, userID, originalURL string) {
	t.Helper()

	if len(events) != 1 {
		t.Fatalf("events count = %d, want 1", len(events))
	}
	if events[0].Timestamp == 0 {
		t.Error("event timestamp is zero")
	}
	if events[0].Action != action {
		t.Errorf("event action = %q, want %q", events[0].Action, action)
	}
	if events[0].UserID != userID {
		t.Errorf("event user id = %q, want %q", events[0].UserID, userID)
	}
	if events[0].URL != originalURL {
		t.Errorf("event URL = %q, want %q", events[0].URL, originalURL)
	}
}
