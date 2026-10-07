package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GLEZH/linkshrtservice/internal/auth"
	"github.com/GLEZH/linkshrtservice/internal/repository"
	"go.uber.org/zap"
)

func BenchmarkHandler_ShortenURLJSON(b *testing.B) {
	handlers := New(
		"http://localhost:8080",
		repository.NewURLStorage(),
		zap.NewNop().Sugar(),
		nil,
	)
	body := `{"url":"https://example.com/articles/benchmark"}`
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
		request = request.WithContext(auth.WithUserID(request.Context(), "user-id"))
		recorder := httptest.NewRecorder()
		handlers.ShortenURLJSON(recorder, request)
	}
}
