package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/GLEZH/linkshrtservice/internal/entity"
)

type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileURLStorage struct {
	*URLStorage
	mu       sync.Mutex
	filePath string
	records  []record
}

func NewFileURLStorage(filePath string, storage *URLStorage) (*FileURLStorage, error) {
	fileStorage := &FileURLStorage{
		URLStorage: storage,
		filePath:   filePath,
	}

	if filePath == "" {
		return fileStorage, nil
	}

	records, err := readRecords(filePath)
	if err != nil {
		return nil, err
	}

	for _, rec := range records {
		storage.restore(entity.URL{ID: rec.ShortURL, OriginalURL: rec.OriginalURL})
	}
	fileStorage.records = records

	return fileStorage, nil
}

func (s *FileURLStorage) Save(url entity.URL) (entity.URL, error) {
	savedURL, err := s.URLStorage.Save(url)
	if err != nil {
		return entity.URL{}, err
	}

	if s.filePath == "" {
		return savedURL, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append(s.records, record{
		UUID:        savedURL.ID,
		ShortURL:    savedURL.ID,
		OriginalURL: savedURL.OriginalURL,
	})

	if err := writeRecords(s.filePath, s.records); err != nil {
		return entity.URL{}, err
	}

	return savedURL, nil
}

func (s *FileURLStorage) SaveBatch(urls []entity.URL) ([]entity.URL, error) {
	savedURLs, err := s.URLStorage.SaveBatch(urls)
	if err != nil {
		return nil, err
	}

	if s.filePath == "" {
		return savedURLs, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, savedURL := range savedURLs {
		s.records = append(s.records, record{
			UUID:        savedURL.ID,
			ShortURL:    savedURL.ID,
			OriginalURL: savedURL.OriginalURL,
		})
	}

	if err := writeRecords(s.filePath, s.records); err != nil {
		return nil, err
	}

	return savedURLs, nil
}

func (s *FileURLStorage) Close() error {
	return nil
}

func readRecords(filename string) ([]record, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read file: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var records []record
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("unmarshal records: %w", err)
	}

	return records, nil
}

func writeRecords(filename string, records []record) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal records: %w", err)
	}

	dir := filepath.Dir(filename)
	tmpFile, err := os.CreateTemp(dir, ".short-url-db-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	tmpName := tmpFile.Name()
	if _, err = tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write temp file: %w", err)
	}

	if err = tmpFile.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close temp file: %w", err)
	}

	if err = os.Chmod(tmpName, 0666); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("chmod temp file: %w", err)
	}

	if err = os.Rename(tmpName, filename); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}
