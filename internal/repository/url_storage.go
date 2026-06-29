package repository

import (
	"strconv"
	"sync"

	"github.com/GLEZH/linkshrtservice/internal/entity"
)

type URLStorage struct {
	mu     sync.RWMutex
	nextID int
	urls   map[string]entity.URL
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		urls: make(map[string]entity.URL),
	}
}

func (s *URLStorage) Save(url entity.URL) (entity.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	id := strconv.Itoa(s.nextID)
	url.ID = id
	s.urls[id] = url

	return url, nil
}

func (s *URLStorage) Get(id string) (entity.URL, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[id]
	return url, ok
}
