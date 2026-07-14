package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestWithLogging(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	sugar := zap.New(core).Sugar()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})

	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	recorder := httptest.NewRecorder()

	WithLogging(handler, sugar).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusCreated)
	}

	if logs.Len() != 1 {
		t.Fatalf("logs count = %d, want 1", logs.Len())
	}

	entry := logs.All()[0]
	if entry.Level != zapcore.InfoLevel {
		t.Errorf("log level = %s, want %s", entry.Level, zapcore.InfoLevel)
	}

	fields := entry.ContextMap()
	if fields["uri"] != "/test" {
		t.Errorf("uri field = %v, want %q", fields["uri"], "/test")
	}

	if fields["method"] != http.MethodPost {
		t.Errorf("method field = %v, want %q", fields["method"], http.MethodPost)
	}

	if fields["status"] != int64(http.StatusCreated) {
		t.Errorf("status field = %v, want %d", fields["status"], http.StatusCreated)
	}

	if fields["size"] != int64(2) {
		t.Errorf("size field = %v, want 2", fields["size"])
	}

	if _, ok := fields["duration"]; !ok {
		t.Error("duration field is missing")
	}
}
