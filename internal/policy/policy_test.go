package policy_test

import (
	"strings"
	"testing"

	"github.com/driftlog/internal/diff"
	"github.com/driftlog/internal/policy"
)

func makeResult(id string, status diff.Status, changes []diff.Change, cloud map[string]interface{}) diff.Result {
	return diff.Result{
		ResourceID:      id,
		ResourceType:    "aws_instance",
		Status:          status,
		Changes:         changes,
		CloudAttributes: cloud,
	}
}

func TestEvaluate_NoDrift_NoPolicyViolations(t *testing.T) {
	results := []diff.Result{
		makeResult("i-001", diff.StatusMatch, nil, map[string]interface{}{"instance_type": "t2.micro"}),
	}
	report := policy.Evaluate(results, policy.DefaultRules())
	if len(report.Violations) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(report.Violations))
	}
	if report.Passed != 2 {
		t.Errorf("expected 2 passed checks, got %d", report.Passed)
	}
}

func TestEvaluate_DriftedResource_ReturnsError(t *testing.T) {
	results := []diff.Result{
		makeResult("i-002", diff.StatusDrifted, []diff.Change{{Attribute: "instance_type"}}, nil),
	}
	report := policy.Evaluate(results, []policy.Rule{policy.NoDriftAllowed()})
	if len(report.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(report.Violations))
	}
	if report.Violations[0].Severity != "error" {
		t.Errorf("expected severity error, got %s", report.Violations[0].Severity)
	}
	if !report.HasErrors() {
		t.Error("expected HasErrors to be true")
	}
}

func TestEvaluate_OnlyInCloud_ReturnsWarn(t *testing.T) {
	results := []diff.Result{
		makeResult("i-003", diff.StatusOnlyInCloud, nil, map[string]interface{}{}),
	}
	report := policy.Evaluate(results, []policy.Rule{policy.NoUnmanagedResources()})
	if len(report.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(report.Violations))
	}
	if report.Violations[0].Severity != "warn" {
		t.Errorf("expected severity warn, got %s", report.Violations[0].Severity)
	}
	if report.HasErrors() {
		t.Error("expected HasErrors to be false for warn-only violations")
	}
}

func TestRequireAttribute_MissingAttribute(t *testing.T) {
	results := []diff.Result{
		makeResult("i-004", diff.StatusDrifted, []diff.Change{{}}, map[string]interface{}{}),
	}
	report := policy.Evaluate(results, []policy.Rule{policy.RequireAttribute("owner")})
	if len(report.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(report.Violations))
	}
	if !strings.Contains(report.Violations[0].Message, "owner") {
		t.Errorf("expected message to mention 'owner', got: %s", report.Violations[0].Message)
	}
}

func TestFormatReport_NoViolations(t *testing.T) {
	report := policy.Report{Passed: 4}
	out := policy.FormatReport(report)
	if !strings.Contains(out, "passed") {
		t.Errorf("expected 'passed' in output, got: %s", out)
	}
}

func TestFormatReport_WithViolations(t *testing.T) {
	report := policy.Report{
		Passed: 1,
		Failed: 1,
		Violations: []policy.Violation{
			{Rule: "no-drift-allowed", Severity: "error", ResourceID: "i-005", Message: "drift detected"},
		},
	}
	out := policy.FormatReport(report)
	if !strings.Contains(out, "i-005") {
		t.Errorf("expected resource ID in output, got: %s", out)
	}
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR severity in output, got: %s", out)
	}
}
