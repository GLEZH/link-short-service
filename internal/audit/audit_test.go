package audit

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileObserver_Notify(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "audit.log")
	observer := NewFileObserver(filePath)
	events := []Event{
		{Timestamp: 1, Action: ActionShorten, UserID: "user-id", URL: "https://example.com/first"},
		{Timestamp: 2, Action: ActionFollow, URL: "https://example.com/second"},
	}

	for _, event := range events {
		if err := observer.Notify(context.Background(), event); err != nil {
			t.Fatalf("Notify() error = %v", err)
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != len(events) {
		t.Fatalf("audit lines = %d, want %d", len(lines), len(events))
	}

	for i, line := range lines {
		var event Event
		if err = json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if event != events[i] {
			t.Errorf("event = %+v, want %+v", event, events[i])
		}
	}
}

func TestHTTPObserver_Notify(t *testing.T) {
	want := Event{Timestamp: 1, Action: ActionShorten, UserID: "user-id", URL: "https://example.com"}
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want %q", r.Method, http.MethodPost)
		}
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
		}

		var event Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("Decode() error = %v", err)
		}
		if event != want {
			t.Errorf("event = %+v, want %+v", event, want)
		}

		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})}

	observer := NewHTTPObserver("http://audit.example.com", client)
	if err := observer.Notify(context.Background(), want); err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
}

func TestPublisher_NotifiesAllObservers(t *testing.T) {
	want := Event{Timestamp: 1, Action: ActionFollow, URL: "https://example.com"}
	recorder := &recordingObserver{}
	publisher := NewPublisher(nil)
	publisher.Subscribe(observerFunc(func(context.Context, Event) error {
		return errors.New("notify failed")
	}))
	publisher.Subscribe(recorder)

	publisher.Notify(context.Background(), want)

	if len(recorder.events) != 1 {
		t.Fatalf("events count = %d, want 1", len(recorder.events))
	}
	if recorder.events[0] != want {
		t.Errorf("event = %+v, want %+v", recorder.events[0], want)
	}
}

type observerFunc func(context.Context, Event) error

func (f observerFunc) Notify(ctx context.Context, event Event) error {
	return f(ctx, event)
}

type recordingObserver struct {
	events []Event
}

func (o *recordingObserver) Notify(_ context.Context, event Event) error {
	o.events = append(o.events, event)
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
