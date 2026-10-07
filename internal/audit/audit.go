package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"go.uber.org/zap"
)

// Action identifies an audited operation
type Action string

const (
	// ActionShorten records short URL creation
	ActionShorten Action = "shorten"
	// ActionFollow records a short URL redirect
	ActionFollow Action = "follow"
)

// Event describes one audit event
type Event struct {
	Timestamp int64  `json:"ts"`
	Action    Action `json:"action"`
	UserID    string `json:"user_id,omitempty"`
	URL       string `json:"url"`
}

// Observer receives audit events
type Observer interface {
	Notify(context.Context, Event) error
}

// Publisher sends events to subscribed observers
type Publisher struct {
	mu        sync.RWMutex
	observers []Observer
	log       *zap.SugaredLogger
}

// NewPublisher creates an audit event publisher
func NewPublisher(log *zap.SugaredLogger) *Publisher {
	return &Publisher{log: log}
}

// Subscribe adds an observer to the publisher
func (p *Publisher) Subscribe(observer Observer) {
	if observer == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.observers = append(p.observers, observer)
}

// Notify sends an event to every observer
func (p *Publisher) Notify(ctx context.Context, event Event) {
	p.mu.RLock()
	observers := append([]Observer(nil), p.observers...)
	p.mu.RUnlock()

	for _, observer := range observers {
		if err := observer.Notify(ctx, event); err != nil && p.log != nil {
			p.log.Infow("send audit event failed", "error", err)
		}
	}
}

// FileObserver appends audit events to a file
type FileObserver struct {
	filePath string
	mu       sync.Mutex
}

// NewFileObserver creates a file audit observer
func NewFileObserver(filePath string) *FileObserver {
	return &FileObserver{filePath: filePath}
}

// Notify appends an event as one JSON line
func (o *FileObserver) Notify(_ context.Context, event Event) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	file, err := os.OpenFile(o.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("open audit file: %w", err)
	}
	defer file.Close()

	if err = json.NewEncoder(file).Encode(event); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}

	return nil
}

// HTTPObserver posts audit events to a remote server
type HTTPObserver struct {
	url    string
	client *http.Client
}

// NewHTTPObserver creates a remote audit observer
func NewHTTPObserver(url string, client *http.Client) *HTTPObserver {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPObserver{url: url, client: client}
}

// Notify posts an event in JSON format
func (o *HTTPObserver) Notify(ctx context.Context, event Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, o.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create audit request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := o.client.Do(request)
	if err != nil {
		return fmt.Errorf("send audit request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("audit server returned status %d", response.StatusCode)
	}

	return nil
}
