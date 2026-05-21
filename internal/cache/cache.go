package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry holds a cached snapshot of cloud resources with metadata.
type Entry struct {
	FetchedAt time.Time              `json:"fetched_at"`
	Resources map[string]interface{} `json:"resources"`
}

// IsExpired reports whether the entry is older than ttl.
func (e *Entry) IsExpired(ttl time.Duration) bool {
	return time.Since(e.FetchedAt) > ttl
}

// Store persists an entry to a JSON file at the given path.
func Store(path string, entry *Entry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("cache: create directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cache: create file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entry); err != nil {
		return fmt.Errorf("cache: encode: %w", err)
	}
	return nil
}

// Load reads a cache entry from the given path.
// Returns (nil, nil) when the file does not exist.
func Load(path string) (*Entry, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return fmt.Errorf("cache: open file: %w", err)
	}
	defer f.Close()
	var entry Entry
	if err := json.NewDecoder(f).Decode(&entry); err != nil {
		return nil, fmt.Errorf("cache: decode: %w", err)
	}
	return &entry, nil
}
