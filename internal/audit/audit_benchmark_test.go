package audit

import (
	"context"
	"testing"
)

func BenchmarkPublisher_Notify(b *testing.B) {
	publisher := NewPublisher(nil)
	publisher.Subscribe(observerFunc(func(context.Context, Event) error {
		return nil
	}))
	event := Event{
		Timestamp: 1,
		Action:    ActionShorten,
		UserID:    "user-id",
		URL:       "https://example.com/articles/benchmark",
	}
	b.ReportAllocs()

	for b.Loop() {
		publisher.Notify(context.Background(), event)
	}

	b.StopTimer()
	if err := publisher.Close(); err != nil {
		b.Fatal(err)
	}
}
