package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GLEZH/linkshrtservice/internal/handler"
	"github.com/GLEZH/linkshrtservice/internal/repository"
	"go.uber.org/zap"
)

type testDatabase struct {
	err error
}

func (d testDatabase) Ping(ctx context.Context) error {
	return d.err
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	storage, err := repository.New("")
	if err != nil {
		t.Fatalf("repository.New() error = %v", err)
	}
	handlers := handler.New("http://localhost:8080", storage, zap.NewNop().Sugar(), testDatabase{})
	return newRouter(handlers, zap.NewNop().Sugar())
}

func newTestRouterWithDatabase(t *testing.T, db handler.Database) http.Handler {
	t.Helper()

	storage, err := repository.New("")
	if err != nil {
		t.Fatalf("repository.New() error = %v", err)
	}
	handlers := handler.New("http://localhost:8080", storage, zap.NewNop().Sugar(), db)
	return newRouter(handlers, zap.NewNop().Sugar())
}

func TestRouter(t *testing.T) {
	t.Run("shorten and expand", func(t *testing.T) {
		router := newTestRouter(t)
		originalURL := "http://example.com"

		shortenRequest := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
		shortenRecorder := httptest.NewRecorder()
		router.ServeHTTP(shortenRecorder, shortenRequest)

		if shortenRecorder.Code != http.StatusCreated {
			t.Fatalf("shorten status code = %d, want %d", shortenRecorder.Code, http.StatusCreated)
		}

		body, err := io.ReadAll(shortenRecorder.Result().Body)
		if err != nil {
			t.Fatalf("io.ReadAll() error = %v", err)
		}

		shortURL := string(body)
		if shortURL == "" {
			t.Fatal("short url is empty")
		}

		expandPath := strings.TrimPrefix(shortURL, "http://localhost:8080")
		expandRequest := httptest.NewRequest(http.MethodGet, expandPath, nil)
		expandRecorder := httptest.NewRecorder()
		router.ServeHTTP(expandRecorder, expandRequest)

		if expandRecorder.Code != http.StatusTemporaryRedirect {
			t.Errorf("expand status code = %d, want %d", expandRecorder.Code, http.StatusTemporaryRedirect)
		}

		if location := expandRecorder.Header().Get("Location"); location != originalURL {
			t.Errorf("Location = %q, want %q", location, originalURL)
		}
	})

	t.Run("json shorten and expand", func(t *testing.T) {
		router := newTestRouter(t)
		originalURL := "https://practicum.yandex.ru"

		shortenRequest := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"`+originalURL+`"}`))
		shortenRequest.Header.Set("Content-Type", "application/json")
		shortenRecorder := httptest.NewRecorder()
		router.ServeHTTP(shortenRecorder, shortenRequest)

		if shortenRecorder.Code != http.StatusCreated {
			t.Fatalf("shorten status code = %d, want %d", shortenRecorder.Code, http.StatusCreated)
		}

		contentType := shortenRecorder.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
		}

		var response struct {
			Result string `json:"result"`
		}
		if err := json.NewDecoder(shortenRecorder.Result().Body).Decode(&response); err != nil {
			t.Fatalf("json decode error = %v", err)
		}

		if response.Result == "" {
			t.Fatal("result is empty")
		}

		expandPath := strings.TrimPrefix(response.Result, "http://localhost:8080")
		expandRequest := httptest.NewRequest(http.MethodGet, expandPath, nil)
		expandRecorder := httptest.NewRecorder()
		router.ServeHTTP(expandRecorder, expandRequest)

		if expandRecorder.Code != http.StatusTemporaryRedirect {
			t.Errorf("expand status code = %d, want %d", expandRecorder.Code, http.StatusTemporaryRedirect)
		}

		if location := expandRecorder.Header().Get("Location"); location != originalURL {
			t.Errorf("Location = %q, want %q", location, originalURL)
		}
	})

	t.Run("json shorten returns gzip response", func(t *testing.T) {
		router := newTestRouter(t)
		originalURL := "https://practicum.yandex.ru"

		request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"`+originalURL+`"}`))
		request.Header.Set("Accept-Encoding", "gzip")
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusCreated)
		}

		if encoding := recorder.Header().Get("Content-Encoding"); encoding != "gzip" {
			t.Fatalf("Content-Encoding = %q, want %q", encoding, "gzip")
		}

		gz, err := gzip.NewReader(recorder.Result().Body)
		if err != nil {
			t.Fatalf("gzip.NewReader() error = %v", err)
		}
		defer gz.Close()

		var response struct {
			Result string `json:"result"`
		}
		if err = json.NewDecoder(gz).Decode(&response); err != nil {
			t.Fatalf("json decode error = %v", err)
		}

		if response.Result == "" {
			t.Fatal("result is empty")
		}
	})

	t.Run("json shorten accepts gzip request", func(t *testing.T) {
		router := newTestRouter(t)
		originalURL := "https://practicum.yandex.ru"
		body := gzipBody(t, `{"url":"`+originalURL+`"}`)

		request := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
		request.Header.Set("Content-Encoding", "gzip")
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusCreated)
		}

		var response struct {
			Result string `json:"result"`
		}
		if err := json.NewDecoder(recorder.Result().Body).Decode(&response); err != nil {
			t.Fatalf("json decode error = %v", err)
		}

		if response.Result == "" {
			t.Fatal("result is empty")
		}
	})

	t.Run("text plain response is not compressed", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com"))
		request.Header.Set("Accept-Encoding", "gzip")
		recorder := httptest.NewRecorder()

		newTestRouter(t).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusCreated)
		}

		if encoding := recorder.Header().Get("Content-Encoding"); encoding != "" {
			t.Errorf("Content-Encoding = %q, want empty", encoding)
		}
	})

	t.Run("json shorten with bad json is bad request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{`))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		newTestRouter(t).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("ping with database connection is ok", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/ping", nil)
		recorder := httptest.NewRecorder()

		newTestRouterWithDatabase(t, testDatabase{}).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusOK)
		}
	})

	t.Run("ping without database connection is internal error", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/ping", nil)
		recorder := httptest.NewRecorder()

		newTestRouterWithDatabase(t, testDatabase{err: errors.New("ping failed")}).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusInternalServerError)
		}
	})

	t.Run("get root is bad request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()

		newTestRouter(t).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("post wrong path is bad request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		recorder := httptest.NewRecorder()

		newTestRouter(t).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("post id is bad request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/abc123", nil)
		recorder := httptest.NewRecorder()

		newTestRouter(t).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})
}

func gzipBody(t *testing.T, body string) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(body)); err != nil {
		t.Fatalf("gzip write error = %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close error = %v", err)
	}

	return &buf
}
