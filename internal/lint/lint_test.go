package lint_test

import (
	"testing"

	"github.com/driftlog/internal/diff"
	"github.com/driftlog/internal/lint"
)

func makeResult(id string, status diff.DriftStatus, cloud map[string]string) diff.DriftResult {
	return diff.DriftResult{
		ResourceID:  id,
		Status:      status,
		CloudAttrs:  cloud,
		StateAttrs:  map[string]string{},
		Changes:     nil,
	}
}

func TestRequireTag_Present(t *testing.T) {
	r := makeResult("i-123", diff.StatusDrifted, map[string]string{"Environment": "prod"})
	violations := lint.RequireTag("Environment")(r)
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}
}

func TestRequireTag_Missing(t *testing.T) {
	r := makeResult("i-456", diff.StatusDrifted, map[string]string{"Name": "web"})
	violations := lint.RequireTag("Environment")(r)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Severity != lint.SeverityError {
		t.Errorf("expected error severity, got %s", violations[0].Severity)
	}
}

func TestRequireTag_SkipsOnlyInState(t *testing.T) {
	r := makeResult("i-789", diff.StatusOnlyInState, map[string]string{})
	violations := lint.RequireTag("Environment")(r)
	if len(violations) != 0 {
		t.Fatalf("expected no violations for OnlyInState, got %d", len(violations))
	}
}

func TestDisallowAttributeValue_Match(t *testing.T) {
	r := makeResult("bucket-1", diff.StatusDrifted, map[string]string{"acl": "public-read"})
	violations := lint.DisallowAttributeValue("acl", "public-read")(r)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Severity != lint.SeverityWarn {
		t.Errorf("expected warn severity, got %s", violations[0].Severity)
	}
}

func TestDisallowAttributeValue_NoMatch(t *testing.T) {
	r := makeResult("bucket-2", diff.StatusDrifted, map[string]string{"acl": "private"})
	violations := lint.DisallowAttributeValue("acl", "public-read")(r)
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}
}

func TestRun_MultipleRulesAndResults(t *testing.T) {
	results := []diff.DriftResult{
		makeResult("i-1", diff.StatusDrifted, map[string]string{"acl": "public-read"}),
		makeResult("i-2", diff.StatusDrifted, map[string]string{"Environment": "prod"}),
	}
	rules := []lint.Rule{
		lint.RequireTag("Environment"),
		lint.DisallowAttributeValue("acl", "public-read"),
	}
	violations := lint.Run(results, rules)
	// i-1: missing Environment tag + disallowed acl => 2 violations
	// i-2: has Environment, acl not present => 0 violations
	if len(violations) != 2 {
		t.Errorf("expected 2 violations, got %d", len(violations))
	}
}

func TestRun_NoRules(t *testing.T) {
	results := []diff.DriftResult{
		makeResult("i-1", diff.StatusDrifted, map[string]string{}),
	}
	violations := lint.Run(results, nil)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations with no rules, got %d", len(violations))
	}
}
