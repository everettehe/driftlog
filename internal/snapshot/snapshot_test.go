package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/driftlog/internal/diff"
	"github.com/yourorg/driftlog/internal/snapshot"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "snapshot-test-*")
	if err != nil {
		t.Fatalf("tempDir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func makeResults(statuses ...string) []diff.DriftResult {
	results := make([]diff.DriftResult, len(statuses))
	for i, s := range statuses {
		results[i] = diff.DriftResult{
			ResourceID:   fmt.Sprintf("res-%d", i),
			ResourceType: "aws_instance",
			Status:       s,
		}
	}
	return results
}

func TestSave_CreatesFile(t *testing.T) {
	dir := tempDir(t)
	results := []diff.DriftResult{
		{ResourceID: "i-123", ResourceType: "aws_instance", Status: "drifted"},
	}

	snap, err := snapshot.Save(dir, results, "test-label")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	path := filepath.Join(dir, snap.ID+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", path)
	}
	if snap.Label != "test-label" {
		t.Errorf("label: got %q, want %q", snap.Label, "test-label")
	}
}

func TestLoad_RoundTrip(t *testing.T) {
	dir := tempDir(t)
	results := []diff.DriftResult{
		{ResourceID: "bucket-1", ResourceType: "aws_s3_bucket", Status: "ok"},
	}

	saved, err := snapshot.Save(dir, results, "")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := snapshot.Load(filepath.Join(dir, saved.ID+".json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.ID != saved.ID {
		t.Errorf("ID mismatch: got %q, want %q", loaded.ID, saved.ID)
	}
	if len(loaded.Results) != 1 || loaded.Results[0].ResourceID != "bucket-1" {
		t.Errorf("results mismatch: %+v", loaded.Results)
	}
}

func TestLoad_MissingFile_ReturnsNil(t *testing.T) {
	snap, err := snapshot.Load("/nonexistent/path.json")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if snap != nil {
		t.Errorf("expected nil snapshot, got: %+v", snap)
	}
}

func TestDiff_DetectsStatusChanges(t *testing.T) {
	prev := &snapshot.Snapshot{
		Results: []diff.DriftResult{
			{ResourceID: "i-1", ResourceType: "aws_instance", Status: "ok"},
			{ResourceID: "i-2", ResourceType: "aws_instance", Status: "ok"},
		},
	}
	curr := &snapshot.Snapshot{
		Results: []diff.DriftResult{
			{ResourceID: "i-1", ResourceType: "aws_instance", Status: "drifted"},
			{ResourceID: "i-2", ResourceType: "aws_instance", Status: "ok"},
			{ResourceID: "i-3", ResourceType: "aws_instance", Status: "missing"},
		},
	}

	changes := snapshot.Diff(prev, curr)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}

	if changes[0].ResourceID != "i-1" || changes[0].PrevStatus != "ok" || changes[0].CurrStatus != "drifted" {
		t.Errorf("unexpected change[0]: %+v", changes[0])
	}
	if changes[1].ResourceID != "i-3" || changes[1].PrevStatus != "new" {
		t.Errorf("unexpected change[1]: %+v", changes[1])
	}
}
