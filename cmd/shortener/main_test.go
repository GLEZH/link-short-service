package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GLEZH/linkshrtservice/internal/handler"
	"github.com/GLEZH/linkshrtservice/internal/repository"
	"go.uber.org/zap"
)

func newTestRouter() http.Handler {
	storage := repository.NewURLStorage()
	handlers := handler.New("http://localhost:8080", storage)
	return newRouter(handlers, zap.NewNop().Sugar())
}

func TestRouter(t *testing.T) {
	t.Run("shorten and expand", func(t *testing.T) {
		router := newTestRouter()
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
		router := newTestRouter()
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

	t.Run("json shorten with bad json is bad request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{`))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		newTestRouter().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("get root is bad request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()

		newTestRouter().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("post wrong path is bad request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		recorder := httptest.NewRecorder()

		newTestRouter().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("post id is bad request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/abc123", nil)
		recorder := httptest.NewRecorder()

		newTestRouter().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("status code = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})
}
