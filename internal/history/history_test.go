package history_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/driftlog/internal/diff"
	"github.com/example/driftlog/internal/history"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "history.json")
}

func makeResults(statuses ...diff.Status) []diff.DriftResult {
	var out []diff.DriftResult
	for i, s := range statuses {
		out = append(out, diff.DriftResult{
			ResourceID: fmt.Sprintf("res-%d", i),
			Status:     s,
		})
	}
	return out
}

func TestStore_CreatesFile(t *testing.T) {
	p := tempPath(t)
	results := []diff.DriftResult{
		{ResourceID: "i-001", Status: diff.StatusMatch},
		{ResourceID: "i-002", Status: diff.StatusDrifted},
	}
	if err := history.Store(p, results); err != nil {
		t.Fatalf("Store: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestStore_AppendsDriftCount(t *testing.T) {
	p := tempPath(t)
	results := []diff.DriftResult{
		{ResourceID: "i-001", Status: diff.StatusMatch},
		{ResourceID: "i-002", Status: diff.StatusDrifted},
		{ResourceID: "i-003", Status: diff.StatusOnlyInCloud},
	}
	if err := history.Store(p, results); err != nil {
		t.Fatalf("Store: %v", err)
	}
	h, err := history.Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(h.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(h.Entries))
	}
	e := h.Entries[0]
	if e.TotalCount != 3 {
		t.Errorf("TotalCount: want 3, got %d", e.TotalCount)
	}
	if e.DriftCount != 2 {
		t.Errorf("DriftCount: want 2, got %d", e.DriftCount)
	}
}

func TestStore_MultipleRuns(t *testing.T) {
	p := tempPath(t)
	for i := 0; i < 3; i++ {
		if err := history.Store(p, []diff.DriftResult{{ResourceID: "r", Status: diff.StatusMatch}}); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}
	h, _ := history.Load(p)
	if len(h.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(h.Entries))
	}
}

func TestLoad_MissingFile_ReturnsEmpty(t *testing.T) {
	h, err := history.Load("/nonexistent/path/history.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(h.Entries) != 0 {
		t.Errorf("expected empty entries")
	}
}

func TestLoad_InvalidJSON_ReturnsError(t *testing.T) {
	p := tempPath(t)
	os.WriteFile(p, []byte("not-json"), 0o644)
	_, err := history.Load(p)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestTrend_LimitsEntries(t *testing.T) {
	p := tempPath(t)
	for i := 0; i < 5; i++ {
		history.Store(p, []diff.DriftResult{{ResourceID: "r", Status: diff.StatusMatch}})
	}
	h, _ := history.Load(p)
	trend := h.Trend(3)
	if len(trend) != 3 {
		t.Errorf("Trend(3): want 3, got %d", len(trend))
	}
}

func TestTrend_SortedAscending(t *testing.T) {
	h := &history.History{
		Entries: []history.Entry{
			{Timestamp: time.Now().Add(-1 * time.Hour)},
			{Timestamp: time.Now().Add(-3 * time.Hour)},
			{Timestamp: time.Now().Add(-2 * time.Hour)},
		},
	}
	trend := h.Trend(0)
	for i := 1; i < len(trend); i++ {
		if trend[i].Timestamp.Before(trend[i-1].Timestamp) {
			t.Errorf("entries not sorted ascending at index %d", i)
		}
	}
}

// ensure json round-trip preserves timestamps
func TestStore_TimestampRoundTrip(t *testing.T) {
	p := tempPath(t)
	before := time.Now().UTC().Truncate(time.Second)
	history.Store(p, nil)
	h, _ := history.Load(p)
	after := time.Now().UTC()
	ts := h.Entries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v out of range [%v, %v]", ts, before, after)
	}
	// verify JSON serialisability
	if _, err := json.Marshal(h); err != nil {
		t.Errorf("json.Marshal: %v", err)
	}
}
