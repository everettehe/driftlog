package policy_test

import (
	"testing"

	"github.com/driftlog/internal/diff"
	"github.com/driftlog/internal/policy"
)

func TestNoDriftAllowed_MatchStatus_NoViolation(t *testing.T) {
	rule := policy.NoDriftAllowed()
	r := makeResult("i-010", diff.StatusMatch, nil, nil)
	if v := rule.Check(r); v != nil {
		t.Errorf("expected no violation for matching resource, got: %+v", v)
	}
}

func TestNoDriftAllowed_DriftedStatus_Violation(t *testing.T) {
	rule := policy.NoDriftAllowed()
	r := makeResult("i-011", diff.StatusDrifted, []diff.Change{{Attribute: "ami"}}, nil)
	v := rule.Check(r)
	if v == nil {
		t.Fatal("expected violation for drifted resource")
	}
}

func TestNoUnmanagedResources_OnlyInState_NoViolation(t *testing.T) {
	rule := policy.NoUnmanagedResources()
	r := makeResult("i-012", diff.StatusOnlyInState, nil, nil)
	if v := rule.Check(r); v != nil {
		t.Errorf("expected no violation for state-only resource, got: %+v", v)
	}
}

func TestNoUnmanagedResources_OnlyInCloud_Violation(t *testing.T) {
	rule := policy.NoUnmanagedResources()
	r := makeResult("i-013", diff.StatusOnlyInCloud, nil, map[string]interface{}{})
	v := rule.Check(r)
	if v == nil {
		t.Fatal("expected violation for cloud-only resource")
	}
}

func TestRequireAttribute_Present_NoViolation(t *testing.T) {
	rule := policy.RequireAttribute("env")
	r := makeResult("i-014", diff.StatusDrifted, []diff.Change{{}}, map[string]interface{}{"env": "prod"})
	if v := rule.Check(r); v != nil {
		t.Errorf("expected no violation when attribute is present, got: %+v", v)
	}
}

func TestRequireAttribute_MatchStatus_Skipped(t *testing.T) {
	rule := policy.RequireAttribute("env")
	r := makeResult("i-015", diff.StatusMatch, nil, map[string]interface{}{})
	if v := rule.Check(r); v != nil {
		t.Errorf("expected no violation for matching resource, got: %+v", v)
	}
}

func TestDefaultRules_ReturnsTwoRules(t *testing.T) {
	rules := policy.DefaultRules()
	if len(rules) != 2 {
		t.Errorf("expected 2 default rules, got %d", len(rules))
	}
}
