package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
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

	return os.WriteFile(filename, data, 0666)
}
