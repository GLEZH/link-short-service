package repository

import (
	"fmt"
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

func New(filePath string) (*FileURLStorage, error) {
	return NewFileURLStorage(filePath, NewURLStorage())
}

func (s *URLStorage) Save(url entity.URL) (entity.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.saveLocked(url), nil
}

func (s *URLStorage) SaveBatch(urls []entity.URL) ([]entity.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	savedURLs := make([]entity.URL, 0, len(urls))
	for _, url := range urls {
		savedURLs = append(savedURLs, s.saveLocked(url))
	}

	return savedURLs, nil
}

func (s *URLStorage) saveLocked(url entity.URL) entity.URL {
	s.nextID++
	id := strconv.Itoa(s.nextID)
	url.ID = id
	s.urls[id] = url

	return url
}

func (s *URLStorage) Get(id string) (entity.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[id]
	if !ok {
		return entity.URL{}, fmt.Errorf("%w: id %s", entity.ErrURLNotFound, id)
	}

	return url, nil
}

func (s *URLStorage) restore(url entity.URL) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[url.ID] = url
	if id, err := strconv.Atoi(url.ID); err == nil && id > s.nextID {
		s.nextID = id
	}
}
