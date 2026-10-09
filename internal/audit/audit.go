package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
)

const (
	observerQueueSize = 256
	httpClientTimeout = 10 * time.Second
	retryWaitMin      = 100 * time.Millisecond
	retryWaitMax      = time.Second
	retryMax          = 3
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

type message struct {
	ctx   context.Context
	event Event
}

type subscription struct {
	observer Observer
	events   chan message
}

// Publisher sends events to subscribed observers
type Publisher struct {
	mu            sync.RWMutex
	subscriptions []*subscription
	log           *zap.SugaredLogger
	closed        bool
	closeOnce     sync.Once
	closeErr      error
	workers       sync.WaitGroup
}

var _ io.Closer = (*Publisher)(nil)

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
	if p.closed {
		p.mu.Unlock()
		return
	}

	subscription := &subscription{
		observer: observer,
		events:   make(chan message, observerQueueSize),
	}
	p.subscriptions = append(p.subscriptions, subscription)
	p.workers.Add(1)
	p.mu.Unlock()

	go p.run(subscription)
}

// Notify sends an event to every observer
func (p *Publisher) Notify(ctx context.Context, event Event) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return
	}

	msg := message{ctx: context.WithoutCancel(ctx), event: event}
	for _, subscription := range p.subscriptions {
		select {
		case subscription.events <- msg:
		default:
			if p.log != nil {
				p.log.Infow("audit event queue is full", "action", event.Action)
			}
		}
	}
}

// Close sends queued events and closes observers
func (p *Publisher) Close() error {
	p.closeOnce.Do(func() {
		p.mu.Lock()
		p.closed = true
		for _, subscription := range p.subscriptions {
			close(subscription.events)
		}
		p.mu.Unlock()

		p.workers.Wait()

		var closeErrors []error
		for _, subscription := range p.subscriptions {
			closer, ok := subscription.observer.(io.Closer)
			if !ok {
				continue
			}
			if err := closer.Close(); err != nil {
				closeErrors = append(closeErrors, err)
			}
		}
		p.closeErr = errors.Join(closeErrors...)
	})

	return p.closeErr
}

func (p *Publisher) run(subscription *subscription) {
	defer p.workers.Done()

	for msg := range subscription.events {
		if err := subscription.observer.Notify(msg.ctx, msg.event); err != nil && p.log != nil {
			p.log.Infow("send audit event failed", "error", err)
		}
	}
}

// FileObserver appends audit events to a file
type FileObserver struct {
	file    *os.File
	encoder *json.Encoder
	mu      sync.Mutex
}

var _ io.Closer = (*FileObserver)(nil)

// NewFileObserver creates a file audit observer
func NewFileObserver(filePath string) (*FileObserver, error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}

	return &FileObserver{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// Notify appends an event as one JSON line
func (o *FileObserver) Notify(_ context.Context, event Event) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := o.encoder.Encode(event); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}

	return nil
}

// Close closes the audit file
func (o *FileObserver) Close() error {
	return o.file.Close()
}

// HTTPObserver posts audit events to a remote server
type HTTPObserver struct {
	url    string
	client *http.Client
}

// NewHTTPObserver creates a remote audit observer
func NewHTTPObserver(url string, client *http.Client) *HTTPObserver {
	if client == nil {
		client = &http.Client{}
	}

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = client
	retryClient.Logger = nil
	retryClient.RetryWaitMin = retryWaitMin
	retryClient.RetryWaitMax = retryWaitMax
	retryClient.RetryMax = retryMax

	standardClient := retryClient.StandardClient()
	standardClient.Timeout = httpClientTimeout

	return &HTTPObserver{url: url, client: standardClient}
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
