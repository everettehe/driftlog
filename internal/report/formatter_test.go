package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/driftlog/internal/diff"
)

func makeDriftResult(id, rtype string, status diff.DriftStatus, diffs []string) diff.DriftResult {
	return diff.DriftResult{
		ResourceID:   id,
		ResourceType: rtype,
		Status:       status,
		Diffs:        diffs,
	}
}

func TestWriteText_NoDrift(t *testing.T) {
	results := []diff.DriftResult{
		makeDriftResult("i-123", "aws_instance", diff.StatusMatch, nil),
	}
	var buf bytes.Buffer
	if err := WriteText(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "1 total") {
		t.Errorf("expected total count in output, got: %s", out)
	}
	if strings.Contains(out, "[DRIFT]") {
		t.Errorf("did not expect drift entries for matching resource")
	}
}

func TestWriteText_WithDrift(t *testing.T) {
	results := []diff.DriftResult{
		makeDriftResult("i-456", "aws_instance", diff.StatusDrifted, []string{"~ instance_type: t2.micro -> t3.small"}),
		makeDriftResult("bucket-a", "aws_s3_bucket", diff.StatusMissing, nil),
	}
	var buf bytes.Buffer
	if err := WriteText(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[DRIFT]") {
		t.Errorf("expected DRIFT label in output")
	}
	if !strings.Contains(out, "[MISSING]") {
		t.Errorf("expected MISSING label in output")
	}
	if !strings.Contains(out, "instance_type") {
		t.Errorf("expected attribute diff in output")
	}
}

func TestSummarize(t *testing.T) {
	results := []diff.DriftResult{
		makeDriftResult("a", "aws_instance", diff.StatusMatch, nil),
		makeDriftResult("b", "aws_instance", diff.StatusDrifted, nil),
		makeDriftResult("c", "aws_s3_bucket", diff.StatusMissing, nil),
		makeDriftResult("d", "aws_s3_bucket", diff.StatusOrphaned, nil),
	}
	s := summarize(results)
	if s.Total != 4 {
		t.Errorf("expected Total=4, got %d", s.Total)
	}
	if s.Drifted != 1 {
		t.Errorf("expected Drifted=1, got %d", s.Drifted)
	}
	if s.Missing != 1 {
		t.Errorf("expected Missing=1, got %d", s.Missing)
	}
	if s.Orphaned != 1 {
		t.Errorf("expected Orphaned=1, got %d", s.Orphaned)
	}
}
