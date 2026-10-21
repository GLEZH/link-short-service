package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithGzip_DecompressesRequest(t *testing.T) {
	body := gzipData(t, []byte("compressed request"))
	handler := WithGzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ReadAll() error = %v", err)
			return
		}
		if string(data) != "compressed request" {
			t.Errorf("request body = %q, want %q", data, "compressed request")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	request.Header.Set("Content-Encoding", "gzip")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status code = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestWithGzip_RejectsInvalidRequest(t *testing.T) {
	called := false
	handler := WithGzip(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not gzip"))
	request.Header.Set("Content-Encoding", "gzip")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if called {
		t.Error("next handler was called")
	}
}

func TestWithGzip_CompressesResponse(t *testing.T) {
	handler := WithGzip(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Length", "18")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Errorf("status code = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if recorder.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Content-Encoding = %q, want gzip", recorder.Header().Get("Content-Encoding"))
	}
	if recorder.Header().Get("Vary") != "Accept-Encoding" {
		t.Errorf("Vary = %q, want Accept-Encoding", recorder.Header().Get("Vary"))
	}
	if recorder.Header().Get("Content-Length") != "" {
		t.Errorf("Content-Length = %q, want empty", recorder.Header().Get("Content-Length"))
	}

	reader, err := gzip.NewReader(recorder.Body)
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(data) != `{"status":"ok"}` {
		t.Errorf("response body = %q, want %q", data, `{"status":"ok"}`)
	}
}

func TestWithGzip_LeavesUnsupportedResponseUncompressed(t *testing.T) {
	handler := WithGzip(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte("plain response"))
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Header().Get("Content-Encoding") != "" {
		t.Errorf("Content-Encoding = %q, want empty", recorder.Header().Get("Content-Encoding"))
	}
	if recorder.Body.String() != "plain response" {
		t.Errorf("response body = %q, want %q", recorder.Body.String(), "plain response")
	}
}

func TestIsCompressibleContentType(t *testing.T) {
	tests := []struct {
		contentType string
		want        bool
	}{
		{contentType: "application/json", want: true},
		{contentType: "text/html; charset=utf-8", want: true},
		{contentType: "text/plain", want: false},
	}

	for _, test := range tests {
		if got := isCompressibleContentType(test.contentType); got != test.want {
			t.Errorf("isCompressibleContentType(%q) = %t, want %t", test.contentType, got, test.want)
		}
	}
}

func gzipData(t *testing.T, data []byte) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, err := writer.Write(data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return buffer.Bytes()
}
