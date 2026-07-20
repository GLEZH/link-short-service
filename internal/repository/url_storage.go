package repository

import (
	"strconv"
	"sync"

	"github.com/GLEZH/linkshrtservice/internal/entity"
)

type URLStorage struct {
	mu       sync.RWMutex
	nextID   int
	filePath string
	urls     map[string]entity.URL
	records  []record
}

func New(filePath string) (*URLStorage, error) {
	storage := &URLStorage{
		urls: make(map[string]entity.URL),
	}

	if filePath == "" {
		return storage, nil
	}

	records, err := readRecords(filePath)
	if err != nil {
		return nil, err
	}

	for _, rec := range records {
		storage.urls[rec.ShortURL] = entity.URL{ID: rec.ShortURL, OriginalURL: rec.OriginalURL}
		if id, convErr := strconv.Atoi(rec.UUID); convErr == nil && id > storage.nextID {
			storage.nextID = id
		}
	}

	storage.filePath = filePath
	storage.records = records

	return storage, nil
}

func (s *URLStorage) Save(url entity.URL) (entity.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	id := strconv.Itoa(s.nextID)
	url.ID = id
	s.urls[id] = url

	if s.filePath != "" {
		s.records = append(s.records, record{UUID: id, ShortURL: id, OriginalURL: url.OriginalURL})
		if err := writeRecords(s.filePath, s.records); err != nil {
			return entity.URL{}, err
		}
	}

	return url, nil
}

func (s *URLStorage) Get(id string) (entity.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[id]
	if !ok {
		return entity.URL{}, entity.ErrURLNotFound
	}

	return url, nil
}

func (s *URLStorage) Close() error {
	return nil
}
