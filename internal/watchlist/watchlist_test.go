package watchlist_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/driftlog/internal/diff"
	"github.com/driftlog/internal/watchlist"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "watchlist.json")
}

func makeResult(id, resourceType string, status diff.Status) diff.Result {
	return diff.Result{ResourceID: id, ResourceType: resourceType, Status: status}
}

func TestAdd_NewEntry(t *testing.T) {
	wl := &watchlist.Watchlist{}
	added := wl.Add("i-123", "aws_instance", "critical prod box")
	if !added {
		t.Fatal("expected Add to return true for new entry")
	}
	if len(wl.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(wl.Entries))
	}
	if wl.Entries[0].ResourceID != "i-123" {
		t.Errorf("unexpected resource ID: %s", wl.Entries[0].ResourceID)
	}
}

func TestAdd_Duplicate_ReturnsFalse(t *testing.T) {
	wl := &watchlist.Watchlist{}
	wl.Add("i-123", "aws_instance", "first")
	added := wl.Add("i-123", "aws_instance", "second")
	if added {
		t.Fatal("expected Add to return false for duplicate")
	}
	if len(wl.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(wl.Entries))
	}
}

func TestRemove_ExistingEntry(t *testing.T) {
	wl := &watchlist.Watchlist{}
	wl.Add("i-123", "aws_instance", "reason")
	removed := wl.Remove("i-123")
	if !removed {
		t.Fatal("expected Remove to return true")
	}
	if len(wl.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(wl.Entries))
	}
}

func TestRemove_MissingEntry_ReturnsFalse(t *testing.T) {
	wl := &watchlist.Watchlist{}
	if wl.Remove("nonexistent") {
		t.Fatal("expected Remove to return false for missing entry")
	}
}

func TestFilter_ReturnsOnlyWatched(t *testing.T) {
	wl := &watchlist.Watchlist{}
	wl.Add("i-abc", "aws_instance", "watched")

	results := []diff.Result{
		makeResult("i-abc", "aws_instance", diff.StatusDrifted),
		makeResult("i-xyz", "aws_instance", diff.StatusDrifted),
	}
	filtered := wl.Filter(results)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered result, got %d", len(filtered))
	}
	if filtered[0].ResourceID != "i-abc" {
		t.Errorf("unexpected resource ID: %s", filtered[0].ResourceID)
	}
}

func TestFilter_CaseInsensitive(t *testing.T) {
	wl := &watchlist.Watchlist{}
	wl.Add("I-ABC", "aws_instance", "uppercase")
	results := []diff.Result{makeResult("i-abc", "aws_instance", diff.StatusClean)}
	if len(wl.Filter(results)) != 1 {
		t.Error("expected case-insensitive match")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	path := tempPath(t)
	wl := &watchlist.Watchlist{}
	wl.Add("bucket-1", "aws_s3_bucket", "compliance")

	if err := wl.SaveFile(path); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}
	loaded, err := watchlist.LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if len(loaded.Entries) != 1 || loaded.Entries[0].ResourceID != "bucket-1" {
		t.Errorf("round-trip mismatch: %+v", loaded.Entries)
	}
}

func TestLoadFile_MissingFile_ReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	wl, err := watchlist.LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wl.Entries) != 0 {
		t.Errorf("expected empty watchlist, got %d entries", len(wl.Entries))
	}
}

func TestLoadFile_InvalidJSON_ReturnsError(t *testing.T) {
	path := tempPath(t)
	os.WriteFile(path, []byte("not-json"), 0o644)
	_, err := watchlist.LoadFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestEntry_AddedAt_IsSet(t *testing.T) {
	wl := &watchlist.Watchlist{}
	wl.Add("r-1", "aws_instance", "test")
	if wl.Entries[0].AddedAt.IsZero() {
		t.Error("expected AddedAt to be set")
	}
}

// Ensure Watchlist serialises cleanly (smoke test for JSON tags).
func TestWatchlist_JSONRoundTrip(t *testing.T) {
	wl := &watchlist.Watchlist{}
	wl.Add("sg-99", "aws_security_group", "open port")
	data, err := json.Marshal(wl)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var wl2 watchlist.Watchlist
	if err := json.Unmarshal(data, &wl2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(wl2.Entries) != 1 {
		t.Errorf("expected 1 entry after JSON round-trip, got %d", len(wl2.Entries))
	}
}
