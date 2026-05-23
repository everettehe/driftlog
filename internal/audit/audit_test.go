package audit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/your-org/driftlog/internal/audit"
	"github.com/your-org/driftlog/internal/diff"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "audit.jsonl")
}

func makeResults() []diff.Result {
	return []diff.Result{
		{ResourceID: "i-001", ResourceType: "aws_instance", Status: diff.StatusDrifted},
		{ResourceID: "i-002", ResourceType: "aws_instance", Status: diff.StatusClean},
		{ResourceID: "bucket-1", ResourceType: "aws_s3_bucket", Status: diff.StatusOnlyInCloud},
		{ResourceID: "sg-1", ResourceType: "aws_security_group", Status: diff.StatusOnlyInState},
	}
}

func TestAppend_CreatesFile(t *testing.T) {
	p := tempPath(t)
	if err := audit.Append(p, makeResults(), "run-1", "terraform.tfstate", "us-east-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestAppend_MultipleTimes_AppendsEntries(t *testing.T) {
	p := tempPath(t)
	for i := 0; i < 3; i++ {
		if err := audit.Append(p, makeResults(), "run", "state.tfstate", "us-west-2"); err != nil {
			t.Fatalf("append %d failed: %v", i, err)
		}
	}
	entries, err := audit.ReadAll(p)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestReadAll_MissingFile_ReturnsNil(t *testing.T) {
	entries, err := audit.ReadAll("/nonexistent/path/audit.jsonl")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil entries, got %v", entries)
	}
}

func TestAppend_CountsCorrectly(t *testing.T) {
	p := tempPath(t)
	if err := audit.Append(p, makeResults(), "run-42", "main.tfstate", "eu-west-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries, err := audit.ReadAll(p)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.TotalCount != 4 {
		t.Errorf("TotalCount: want 4, got %d", e.TotalCount)
	}
	if e.DriftCount != 1 {
		t.Errorf("DriftCount: want 1, got %d", e.DriftCount)
	}
	if e.CleanCount != 1 {
		t.Errorf("CleanCount: want 1, got %d", e.CleanCount)
	}
	if e.UnmanagedCount != 1 {
		t.Errorf("UnmanagedCount: want 1, got %d", e.UnmanagedCount)
	}
	if e.MissingCount != 1 {
		t.Errorf("MissingCount: want 1, got %d", e.MissingCount)
	}
	if e.RunID != "run-42" {
		t.Errorf("RunID: want run-42, got %s", e.RunID)
	}
	if e.Region != "eu-west-1" {
		t.Errorf("Region: want eu-west-1, got %s", e.Region)
	}
	if e.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
	if e.Timestamp.After(time.Now().Add(time.Second)) {
		t.Error("Timestamp should not be in the future")
	}
}
