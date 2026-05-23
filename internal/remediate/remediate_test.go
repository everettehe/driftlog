package remediate_test

import (
	"strings"
	"testing"

	"github.com/driftlog/internal/diff"
	"github.com/driftlog/internal/remediate"
)

func makeResult(id, rtype string, status diff.Status, changes []diff.Change) diff.Result {
	return diff.Result{
		ResourceID: id,
		ResourceType: rtype,
		Status: status,
		Changes: changes,
	}
}

func TestSuggest_NoDrift(t *testing.T) {
	results := []diff.Result{
		makeResult("i-abc123", "aws_instance", diff.StatusMatch, nil),
	}
	actions := remediate.Suggest(results)
	if len(actions) != 0 {
		t.Errorf("expected 0 actions, got %d", len(actions))
	}
}

func TestSuggest_Drifted(t *testing.T) {
	changes := []diff.Change{{Attribute: "instance_type", StateValue: "t2.micro", CloudValue: "t3.small"}}
	results := []diff.Result{
		makeResult("i-abc123", "aws_instance", diff.StatusDrifted, changes),
	}
	actions := remediate.Suggest(results)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if !strings.Contains(actions[0].Command, "terraform apply") {
		t.Errorf("expected apply command, got: %s", actions[0].Command)
	}
	if !strings.Contains(actions[0].Command, "i-abc123") {
		t.Errorf("expected resource ID in command, got: %s", actions[0].Command)
	}
}

func TestSuggest_OnlyInCloud(t *testing.T) {
	results := []diff.Result{
		makeResult("bucket-xyz", "aws_s3_bucket", diff.StatusOnlyInCloud, nil),
	}
	actions := remediate.Suggest(results)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if !strings.Contains(actions[0].Command, "terraform import") {
		t.Errorf("expected import command, got: %s", actions[0].Command)
	}
}

func TestSuggest_OnlyInState(t *testing.T) {
	results := []diff.Result{
		makeResult("i-gone", "aws_instance", diff.StatusOnlyInState, nil),
	}
	actions := remediate.Suggest(results)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if !strings.Contains(actions[0].Command, "terraform destroy") {
		t.Errorf("expected destroy command, got: %s", actions[0].Command)
	}
}

func TestFormat_NoActions(t *testing.T) {
	out := remediate.Format(nil)
	if !strings.Contains(out, "No remediation") {
		t.Errorf("expected no-action message, got: %s", out)
	}
}

func TestFormat_WithActions(t *testing.T) {
	actions := []remediate.Action{
		{
			ResourceID: "i-abc",
			ResourceType: "aws_instance",
			Command: "terraform apply -target=i-abc",
			Description: "Resource has 1 attribute(s) out of sync.",
		},
	}
	out := remediate.Format(actions)
	if !strings.Contains(out, "i-abc") {
		t.Errorf("expected resource ID in output")
	}
	if !strings.Contains(out, "terraform apply") {
		t.Errorf("expected command in output")
	}
	if !strings.Contains(out, "Remediation suggestions (1)") {
		t.Errorf("expected summary header in output")
	}
}
