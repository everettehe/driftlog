package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/driftlog/internal/diff"
	"github.com/user/driftlog/internal/output"
)

func makeResult(id, resType string, status diff.DriftStatus) diff.DriftResult {
	return diff.DriftResult{
		ResourceID:   id,
		ResourceType: resType,
		Status:       status,
	}
}

func TestParseFormat_Valid(t *testing.T) {
	cases := []struct {
		input    string
		expected output.Format
	}{
		{"text", output.FormatText},
		{"TEXT", output.FormatText},
		{"", output.FormatText},
		{"json", output.FormatJSON},
		{"JSON", output.FormatJSON},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := output.ParseFormat(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Errorf("got %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestParseFormat_Invalid(t *testing.T) {
	_, err := output.ParseFormat("xml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("error should mention the bad format, got: %v", err)
	}
}

func TestWrite_TextFormat(t *testing.T) {
	results := []diff.DriftResult{
		makeResult("i-123", "aws_instance", diff.StatusMatch),
	}
	var buf bytes.Buffer
	if err := output.Write(&buf, results, output.FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty text output")
	}
}

func TestWrite_JSONFormat(t *testing.T) {
	results := []diff.DriftResult{
		makeResult("bucket-1", "aws_s3_bucket", diff.StatusDrifted),
	}
	var buf bytes.Buffer
	if err := output.Write(&buf, results, output.FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "bucket-1") {
		t.Errorf("expected resource id in JSON output, got: %s", got)
	}
	if !strings.HasPrefix(strings.TrimSpace(got), "[") {
		t.Errorf("expected JSON array output, got: %s", got)
	}
}
