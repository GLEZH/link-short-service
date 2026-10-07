package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/GLEZH/linkshrtservice/internal/entity"
)

func BenchmarkURLStorage_SaveBatch(b *testing.B) {
	ctx := context.Background()
	urls := makeBenchmarkURLs(1000, "user-id")
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage := NewURLStorage()
		if _, err := storage.SaveBatch(ctx, urls); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkURLStorage_GetByUserID(b *testing.B) {
	ctx := context.Background()
	storage := NewURLStorage()
	urls := makeBenchmarkURLs(10000, "user-id")
	if _, err := storage.SaveBatch(ctx, urls); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := storage.GetByUserID(ctx, "user-id"); err != nil {
			b.Fatal(err)
		}
	}
}

func makeBenchmarkURLs(count int, userID string) []entity.URL {
	urls := make([]entity.URL, count)
	for i := range urls {
		urls[i] = entity.URL{
			OriginalURL: fmt.Sprintf("https://example.com/articles/%d", i),
			UserID:      userID,
		}
	}
	return urls
}
