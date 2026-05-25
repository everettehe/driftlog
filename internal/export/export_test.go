package export_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yourorg/driftlog/internal/diff"
	"github.com/yourorg/driftlog/internal/export"
)

func makeResult(id, rtype string, status diff.Status, changes []diff.Change) diff.Result {
	return diff.Result{
		ResourceID:   id,
		ResourceType: rtype,
		Status:       status,
		Changes:      changes,
	}
}

func TestParseFormat_Valid(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"csv", "csv"}, {"JSON", "json"}, {"NDJSON", "ndjson"},
	} {
		f, err := export.ParseFormat(tc.in)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.in, err)
		}
		if string(f) != tc.want {
			t.Errorf("got %q want %q", f, tc.want)
		}
	}
}

func TestParseFormat_Invalid(t *testing.T) {
	_, err := export.ParseFormat("xml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestWrite_CSV_NoChanges(t *testing.T) {
	results := []diff.Result{
		makeResult("i-123", "aws_instance", diff.StatusClean, nil),
	}
	var buf bytes.Buffer
	if err := export.Write(&buf, results, export.FormatCSV); err != nil {
		t.Fatalf("Write: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 row), got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "resource_id") {
		t.Errorf("header missing: %s", lines[0])
	}
}

func TestWrite_CSV_WithChanges(t *testing.T) {
	results := []diff.Result{
		makeResult("i-456", "aws_instance", diff.StatusDrifted, []diff.Change{
			{Attribute: "instance_type", StateValue: "t2.micro", CloudValue: "t3.medium"},
		}),
	}
	var buf bytes.Buffer
	if err := export.Write(&buf, results, export.FormatCSV); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !strings.Contains(buf.String(), "instance_type") {
		t.Errorf("expected attribute in CSV output")
	}
}

func TestWrite_JSON(t *testing.T) {
	results := []diff.Result{
		makeResult("bucket-1", "aws_s3_bucket", diff.StatusClean, nil),
	}
	var buf bytes.Buffer
	if err := export.Write(&buf, results, export.FormatJSON); err != nil {
		t.Fatalf("Write: %v", err)
	}
	var out []diff.Result
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 1 || out[0].ResourceID != "bucket-1" {
		t.Errorf("unexpected output: %+v", out)
	}
}

func TestWrite_NDJSON(t *testing.T) {
	results := []diff.Result{
		makeResult("r-1", "aws_instance", diff.StatusDrifted, nil),
		makeResult("r-2", "aws_s3_bucket", diff.StatusClean, nil),
	}
	var buf bytes.Buffer
	if err := export.Write(&buf, results, export.FormatNDJSON); err != nil {
		t.Fatalf("Write: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 NDJSON lines, got %d", len(lines))
	}
	for _, line := range lines {
		var r diff.Result
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			t.Errorf("invalid NDJSON line: %v", err)
		}
	}
}
