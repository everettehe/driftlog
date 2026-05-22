package redact_test

import (
	"testing"

	"github.com/example/driftlog/internal/diff"
	"github.com/example/driftlog/internal/redact"
)

func makeResult(id string, diffs []diff.AttributeDiff) diff.Result {
	return diff.Result{
		ResourceID:   id,
		ResourceType: "aws_instance",
		Status:       diff.StatusDrifted,
		Diffs:        diffs,
	}
}

func TestApply_NoSensitiveKeys(t *testing.T) {
	results := []diff.Result{
		makeResult("i-001", []diff.AttributeDiff{
			{Key: "instance_type", StateValue: "t2.micro", CloudValue: "t3.micro"},
		}),
	}
	out := redact.Apply(results, []string{})
	if out[0].Diffs[0].StateValue != "t2.micro" {
		t.Errorf("expected value to be unchanged, got %s", out[0].Diffs[0].StateValue)
	}
}

func TestApply_RedactsSensitiveKey(t *testing.T) {
	results := []diff.Result{
		makeResult("i-002", []diff.AttributeDiff{
			{Key: "db_password", StateValue: "hunter2", CloudValue: "supersecret"},
			{Key: "instance_type", StateValue: "t2.micro", CloudValue: "t3.micro"},
		}),
	}
	out := redact.Apply(results, nil)

	passwordDiff := out[0].Diffs[0]
	if passwordDiff.StateValue != "[REDACTED]" {
		t.Errorf("expected StateValue to be redacted, got %s", passwordDiff.StateValue)
	}
	if passwordDiff.CloudValue != "[REDACTED]" {
		t.Errorf("expected CloudValue to be redacted, got %s", passwordDiff.CloudValue)
	}

	instanceDiff := out[0].Diffs[1]
	if instancDiff.StateValue != "t2.micro" {
		t.Errorf("expected non-sensitive value to be unchanged, got %s", instancDiff.StateValue)
	}
}

func TestApply_CaseInsensitiveMatch(t *testing.T) {
	results := []diff.Result{
		makeResult("i-003", []diff.AttributeDiff{
			{Key: "AWS_SECRET_KEY", StateValue: "abc", CloudValue: "xyz"},
		}),
	}
	out := redact.Apply(results, nil)
	if out[0].Diffs[0].StateValue != "[REDACTED]" {
		t.Errorf("expected uppercase sensitive key to be redacted")
	}
}

func TestApply_CustomSensitiveKeys(t *testing.T) {
	results := []diff.Result{
		makeResult("i-004", []diff.AttributeDiff{
			{Key: "internal_code", StateValue: "1234", CloudValue: "5678"},
		}),
	}
	out := redact.Apply(results, []string{"internal_code"})
	if out[0].Diffs[0].StateValue != "[REDACTED]" {
		t.Errorf("expected custom key to be redacted, got %s", out[0].Diffs[0].StateValue)
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	original := []diff.Result{
		makeResult("i-005", []diff.AttributeDiff{
			{Key: "api_key", StateValue: "original", CloudValue: "changed"},
		}),
	}
	redact.Apply(original, nil)
	if original[0].Diffs[0].StateValue != "original" {
		t.Errorf("Apply should not mutate the original slice")
	}
}
