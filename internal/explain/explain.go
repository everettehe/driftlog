// Package explain provides human-readable explanations for detected drift.
package explain

import (
	"fmt"
	"strings"

	"github.com/driftlog/internal/diff"
)

// Severity represents the impact level of a drift.
type Severity string

const (
	SeverityHigh   Severity = "HIGH"
	SeverityMedium Severity = "MEDIUM"
	SeverityLow    Severity = "LOW"
)

// Explanation holds a human-readable description and severity for a drift result.
type Explanation struct {
	ResourceID string
	Severity   Severity
	Summary    string
	Details    []string
}

// highImpactKeys are attribute keys that indicate high-severity drift.
var highImpactKeys = map[string]bool{
	"instance_type": true,
	"ami":           true,
	"subnet_id":     true,
	"vpc_id":        true,
	"iam_role":      true,
	"security_groups": true,
}

// Explain generates an Explanation for a single DriftResult.
func Explain(result diff.DriftResult) Explanation {
	exp := Explanation{
		ResourceID: result.ResourceID,
	}

	switch result.Status {
	case diff.StatusMatch:
		exp.Severity = SeverityLow
		exp.Summary = fmt.Sprintf("Resource %q is in sync with Terraform state.", result.ResourceID)
		return exp

	case diff.StatusOnlyInState:
		exp.Severity = SeverityHigh
		exp.Summary = fmt.Sprintf("Resource %q exists in Terraform state but was not found in the cloud.", result.ResourceID)
		exp.Details = []string{"The resource may have been deleted outside of Terraform.", "Consider running 'terraform plan' to reconcile."}
		return exp

	case diff.StatusOnlyInCloud:
		exp.Severity = SeverityMedium
		exp.Summary = fmt.Sprintf("Resource %q exists in the cloud but is not tracked by Terraform state.", result.ResourceID)
		exp.Details = []string{"The resource may have been created manually.", "Consider importing it with 'terraform import'."}
		return exp
	}

	// StatusDrifted
	severity := SeverityLow
	var details []string
	for _, change := range result.Changes {
		key := strings.ToLower(change.Key)
		if highImpactKeys[key] {
			severity = SeverityHigh
		}
		details = append(details, fmt.Sprintf("  • %s: state=%q cloud=%q", change.Key, change.StateValue, change.CloudValue))
	}
	if severity != SeverityHigh && len(result.Changes) > 2 {
		severity = SeverityMedium
	}

	exp.Severity = severity
	exp.Summary = fmt.Sprintf("Resource %q has %d attribute(s) that differ from Terraform state.", result.ResourceID, len(result.Changes))
	exp.Details = details
	return exp
}

// ExplainAll generates Explanations for a slice of DriftResults.
func ExplainAll(results []diff.DriftResult) []Explanation {
	out := make([]Explanation, 0, len(results))
	for _, r := range results {
		out = append(out, Explain(r))
	}
	return out
}
