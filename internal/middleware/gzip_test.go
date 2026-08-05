package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithGzip_BadGzipRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not gzip"))
	request.Header.Set("Content-Encoding", "gzip")
	recorder := httptest.NewRecorder()

	WithGzip(handler).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestWithGzip_HTMLResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>ok</html>"))
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	recorder := httptest.NewRecorder()

	WithGzip(handler).ServeHTTP(recorder, request)

	if encoding := recorder.Header().Get("Content-Encoding"); encoding != "gzip" {
		t.Fatalf("Content-Encoding = %q, want %q", encoding, "gzip")
	}

	gz, err := gzip.NewReader(recorder.Result().Body)
	if err != nil {
		t.Fatalf("gzip.NewReader() error = %v", err)
	}
	defer gz.Close()

	body, err := io.ReadAll(gz)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}

	if string(body) != "<html>ok</html>" {
		t.Errorf("body = %q, want %q", string(body), "<html>ok</html>")
	}
}
