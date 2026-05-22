package explain_test

import (
	"strings"
	"testing"

	"github.com/driftlog/internal/diff"
	"github.com/driftlog/internal/explain"
)

func makeChange(key, stateVal, cloudVal string) diff.AttributeChange {
	return diff.AttributeChange{Key: key, StateValue: stateVal, CloudValue: cloudVal}
}

func makeResult(id string, status diff.DriftStatus, changes ...diff.AttributeChange) diff.DriftResult {
	return diff.DriftResult{ResourceID: id, Status: status, Changes: changes}
}

func TestExplain_Match(t *testing.T) {
	r := makeResult("i-abc123", diff.StatusMatch)
	exp := explain.Explain(r)

	if exp.Severity != explain.SeverityLow {
		t.Errorf("expected LOW severity, got %s", exp.Severity)
	}
	if !strings.Contains(exp.Summary, "in sync") {
		t.Errorf("expected 'in sync' in summary, got: %s", exp.Summary)
	}
}

func TestExplain_OnlyInState(t *testing.T) {
	r := makeResult("i-missing", diff.StatusOnlyInState)
	exp := explain.Explain(r)

	if exp.Severity != explain.SeverityHigh {
		t.Errorf("expected HIGH severity, got %s", exp.Severity)
	}
	if !strings.Contains(exp.Summary, "not found in the cloud") {
		t.Errorf("unexpected summary: %s", exp.Summary)
	}
	if len(exp.Details) == 0 {
		t.Error("expected non-empty details for OnlyInState")
	}
}

func TestExplain_OnlyInCloud(t *testing.T) {
	r := makeResult("i-untracked", diff.StatusOnlyInCloud)
	exp := explain.Explain(r)

	if exp.Severity != explain.SeverityMedium {
		t.Errorf("expected MEDIUM severity, got %s", exp.Severity)
	}
	if !strings.Contains(exp.Summary, "not tracked by Terraform") {
		t.Errorf("unexpected summary: %s", exp.Summary)
	}
}

func TestExplain_Drifted_HighImpactKey(t *testing.T) {
	r := makeResult("i-drift", diff.StatusDrifted, makeChange("instance_type", "t2.micro", "t3.large"))
	exp := explain.Explain(r)

	if exp.Severity != explain.SeverityHigh {
		t.Errorf("expected HIGH for instance_type drift, got %s", exp.Severity)
	}
	if len(exp.Details) != 1 {
		t.Errorf("expected 1 detail line, got %d", len(exp.Details))
	}
	if !strings.Contains(exp.Details[0], "instance_type") {
		t.Errorf("expected detail to mention 'instance_type', got: %s", exp.Details[0])
	}
}

func TestExplain_Drifted_LowImpact(t *testing.T) {
	r := makeResult("i-minor", diff.StatusDrifted, makeChange("tags", "env=prod", "env=staging"))
	exp := explain.Explain(r)

	if exp.Severity != explain.SeverityLow {
		t.Errorf("expected LOW for tag drift, got %s", exp.Severity)
	}
}

func TestExplainAll_ReturnsOnePerResult(t *testing.T) {
	results := []diff.DriftResult{
		makeResult("i-1", diff.StatusMatch),
		makeResult("i-2", diff.StatusOnlyInState),
		makeResult("i-3", diff.StatusDrifted, makeChange("ami", "ami-old", "ami-new")),
	}
	exps := explain.ExplainAll(results)
	if len(exps) != len(results) {
		t.Errorf("expected %d explanations, got %d", len(results), len(exps))
	}
}
