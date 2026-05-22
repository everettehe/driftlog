package baseline_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/driftlog/internal/baseline"
	"github.com/yourorg/driftlog/internal/diff"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "baseline.json")
}

func makeResults(ids ...string) []diff.DriftResult {
	results := make([]diff.DriftResult, 0, len(ids))
	for _, id := range ids {
		results = append(results, diff.DriftResult{
			ResourceID:   id,
			ResourceType: "aws_instance",
			Status:       diff.StatusDrifted,
		})
	}
	return results
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	path := tempPath(t)
	results := makeResults("i-001", "i-002")

	if err := baseline.Save(path, results); err != nil {
		t.Fatalf("Save: %v", err)
	}

	snap, err := baseline.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if snap == nil {
		t.Fatal("expected snapshot, got nil")
	}
	if len(snap.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(snap.Results))
	}
	if snap.CapturedAt.IsZero() {
		t.Error("expected non-zero CapturedAt")
	}
	if snap.CapturedAt.After(time.Now().Add(time.Second)) {
		t.Error("CapturedAt is in the future")
	}
}

func TestLoad_MissingFile_ReturnsNil(t *testing.T) {
	snap, err := baseline.Load("/nonexistent/path/baseline.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap != nil {
		t.Error("expected nil snapshot for missing file")
	}
}

func TestLoad_InvalidJSON_ReturnsError(t *testing.T) {
	path := tempPath(t)
	_ = os.WriteFile(path, []byte("not-json"), 0o644)

	_, err := baseline.Load(path)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestNewDrifts_NilBaseline_ReturnsAll(t *testing.T) {
	current := makeResults("i-001", "i-002")
	got := baseline.NewDrifts(nil, current)
	if len(got) != 2 {
		t.Errorf("expected 2 results, got %d", len(got))
	}
}

func TestNewDrifts_FiltersKnown(t *testing.T) {
	snap := &baseline.Snapshot{
		Results: makeResults("i-001"),
	}
	current := makeResults("i-001", "i-002", "i-003")

	got := baseline.NewDrifts(snap, current)
	if len(got) != 2 {
		t.Errorf("expected 2 new drifts, got %d", len(got))
	}
	for _, r := range got {
		if r.ResourceID == "i-001" {
			t.Error("i-001 should have been filtered as known drift")
		}
	}
}

func TestNewDrifts_EmptyBaseline_ReturnsAll(t *testing.T) {
	snap := &baseline.Snapshot{Results: []diff.DriftResult{}}
	current := makeResults("i-001")

	got := baseline.NewDrifts(snap, current)
	if len(got) != 1 {
		t.Errorf("expected 1 result, got %d", len(got))
	}
}
