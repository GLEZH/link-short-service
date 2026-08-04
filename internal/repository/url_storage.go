package repository

import (
	"context"
	"strconv"
	"sync"

	"github.com/GLEZH/linkshrtservice/internal/entity"
)

type URLStorage struct {
	mu            sync.RWMutex
	nextID        int
	urls          map[string]entity.URL
	originalIndex map[string]string
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		urls:          make(map[string]entity.URL),
		originalIndex: make(map[string]string),
	}
}

func New(filePath string) (*FileURLStorage, error) {
	return NewFileURLStorage(filePath, NewURLStorage())
}

func (s *URLStorage) Save(ctx context.Context, url entity.URL) (entity.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.originalIndex[url.OriginalURL]; ok {
		existingURL := s.urls[id]
		return existingURL, entity.NewURLAlreadyExistsError(existingURL)
	}

	return s.saveLocked(url), nil
}

func (s *URLStorage) SaveBatch(ctx context.Context, urls []entity.URL) ([]entity.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	savedURLs := make([]entity.URL, 0, len(urls))
	for _, url := range urls {
		savedURLs = append(savedURLs, s.saveLocked(url))
	}

	return savedURLs, nil
}

func (s *URLStorage) saveLocked(url entity.URL) entity.URL {
	if id, ok := s.originalIndex[url.OriginalURL]; ok {
		return s.urls[id]
	}

	s.nextID++
	id := strconv.Itoa(s.nextID)
	url.ID = id
	s.urls[id] = url
	s.originalIndex[url.OriginalURL] = id

	return url
}

func (s *URLStorage) Get(ctx context.Context, id string) (entity.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[id]
	if !ok {
		return entity.URL{}, entity.NewURLNotFoundError(id)
	}

	return url, nil
}

func (s *URLStorage) restore(url entity.URL) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[url.ID] = url
	s.originalIndex[url.OriginalURL] = url.ID
	if id, err := strconv.Atoi(url.ID); err == nil && id > s.nextID {
		s.nextID = id
	}
}
