package repository

import (
	"cmp"
	"context"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/GLEZH/linkshrtservice/internal/entity"
)

// URLStorage stores shortened URLs in memory
type URLStorage struct {
	mu            sync.RWMutex
	nextID        int
	urls          map[string]entity.URL
	originalIndex map[string]string
}

// NewURLStorage creates an empty in-memory storage
func NewURLStorage() *URLStorage {
	return &URLStorage{
		urls:          make(map[string]entity.URL),
		originalIndex: make(map[string]string),
	}
}

// New creates a file-backed URL storage
func New(filePath string) (*FileURLStorage, error) {
	return NewFileURLStorage(filePath, NewURLStorage())
}

// Save stores one URL
func (s *URLStorage) Save(ctx context.Context, url entity.URL) (entity.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.originalIndex[url.OriginalURL]; ok {
		existingURL := s.urls[id]
		return existingURL, entity.NewURLAlreadyExistsError(existingURL)
	}

	return s.saveLocked(url), nil
}

// SaveBatch stores several URLs
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

// Get returns a URL by its short ID
func (s *URLStorage) Get(ctx context.Context, id string) (entity.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[id]
	if !ok {
		return entity.URL{}, entity.NewURLNotFoundError(id)
	}
	if url.IsDeleted {
		return entity.URL{}, entity.NewURLDeletedError(id)
	}

	return url, nil
}

// GetByUserID returns URLs owned by a user
func (s *URLStorage) GetByUserID(ctx context.Context, userID string) ([]entity.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, url := range s.urls {
		if url.UserID == userID {
			count++
		}
	}

	urls := make([]entity.URL, 0, count)
	for _, url := range s.urls {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}

	slices.SortFunc(urls, compareURLIDs)

	return urls, nil
}

func compareURLIDs(a, b entity.URL) int {
	aID, aErr := strconv.Atoi(a.ID)
	bID, bErr := strconv.Atoi(b.ID)
	if aErr != nil || bErr != nil {
		return strings.Compare(a.ID, b.ID)
	}
	return cmp.Compare(aID, bID)
}

// DeleteBatch marks a user's URLs as deleted
func (s *URLStorage) DeleteBatch(ctx context.Context, userID string, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range ids {
		url, ok := s.urls[id]
		if !ok || url.UserID != userID {
			continue
		}
		url.IsDeleted = true
		s.urls[id] = url
	}

	return nil
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
