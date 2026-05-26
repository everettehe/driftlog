package stale_test

import (
	"strings"
	"testing"
	"time"

	"github.com/driftlog/internal/diff"
	"github.com/driftlog/internal/stale"
)

func makeResult(id, rtype string, status diff.Status) diff.DriftResult {
	return diff.DriftResult{
		ResourceID: id,
		ResourceType: rtype,
		Status: status,
	}
}

func TestDetect_NoChanges_ReturnsStale(t *testing.T) {
	prev := []diff.DriftResult{makeResult("i-123", "aws_instance", diff.StatusDrifted)}
	curr := []diff.DriftResult{makeResult("i-123", "aws_instance", diff.StatusDrifted)}

	report := stale.Detect(prev, curr, time.Now().Add(-time.Hour))
	if len(report.Entries) != 1 {
		t.Fatalf("expected 1 stale entry, got %d", len(report.Entries))
	}
	if report.Entries[0].ResourceID != "i-123" {
		t.Errorf("unexpected resource ID: %s", report.Entries[0].ResourceID)
	}
}

func TestDetect_StatusChanged_NotStale(t *testing.T) {
	prev := []diff.DriftResult{makeResult("i-456", "aws_instance", diff.StatusDrifted)}
	curr := []diff.DriftResult{makeResult("i-456", "aws_instance", diff.StatusClean)}

	report := stale.Detect(prev, curr, time.Now().Add(-time.Hour))
	if len(report.Entries) != 0 {
		t.Errorf("expected 0 stale entries, got %d", len(report.Entries))
	}
}

func TestDetect_NewResource_NotStale(t *testing.T) {
	prev := []diff.DriftResult{}
	curr := []diff.DriftResult{makeResult("i-789", "aws_instance", diff.StatusDrifted)}

	report := stale.Detect(prev, curr, time.Now().Add(-time.Hour))
	if len(report.Entries) != 0 {
		t.Errorf("expected 0 stale entries for new resource, got %d", len(report.Entries))
	}
}

func TestDetect_EmptyInputs_ReturnsEmpty(t *testing.T) {
	report := stale.Detect(nil, nil, time.Now())
	if len(report.Entries) != 0 {
		t.Errorf("expected empty report, got %d entries", len(report.Entries))
	}
}

func TestLines_NoEntries(t *testing.T) {
	report := stale.Report{GeneratedAt: time.Now()}
	lines := stale.Lines(report)
	if len(lines) != 1 || !strings.Contains(lines[0], "No stale") {
		t.Errorf("unexpected output: %v", lines)
	}
}

func TestLines_WithEntries(t *testing.T) {
	prev := []diff.DriftResult{makeResult("bucket-1", "aws_s3_bucket", diff.StatusDrifted)}
	curr := []diff.DriftResult{makeResult("bucket-1", "aws_s3_bucket", diff.StatusDrifted)}
	report := stale.Detect(prev, curr, time.Now().Add(-24*time.Hour))

	lines := stale.Lines(report)
	if len(lines) < 2 {
		t.Fatalf("expected header + entries, got %d lines", len(lines))
	}
	if !strings.Contains(lines[1], "bucket-1") {
		t.Errorf("expected resource ID in output, got: %s", lines[1])
	}
}
