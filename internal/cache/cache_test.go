package cache_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/driftlog/internal/cache"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "cache.json")
}

func TestStoreAndLoad_RoundTrip(t *testing.T) {
	path := tempPath(t)
	entry := &cache.Entry{
		FetchedAt: time.Now().Truncate(time.Second),
		Resources: map[string]interface{}{"aws_instance.web": map[string]interface{}{"id": "i-123"}},
	}
	if err := cache.Store(path, entry); err != nil {
		t.Fatalf("Store: %v", err)
	}
	loaded, err := cache.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected entry, got nil")
	}
	if !loaded.FetchedAt.Equal(entry.FetchedAt) {
		t.Errorf("FetchedAt mismatch: got %v, want %v", loaded.FetchedAt, entry.FetchedAt)
	}
}

func TestLoad_MissingFile_ReturnsNil(t *testing.T) {
	entry, err := cache.Load("/nonexistent/path/cache.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry != nil {
		t.Errorf("expected nil entry for missing file, got %+v", entry)
	}
}

func TestLoad_InvalidJSON_ReturnsError(t *testing.T) {
	path := tempPath(t)
	if err := os.WriteFile(path, []byte("not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := cache.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestIsExpired(t *testing.T) {
	old := &cache.Entry{FetchedAt: time.Now().Add(-2 * time.Hour)}
	if !old.IsExpired(time.Hour) {
		t.Error("expected old entry to be expired")
	}
	fresh := &cache.Entry{FetchedAt: time.Now()}
	if fresh.IsExpired(time.Hour) {
		t.Error("expected fresh entry to not be expired")
	}
}
