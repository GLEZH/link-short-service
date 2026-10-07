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
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		publisher.Notify(context.Background(), event)
	}
}
