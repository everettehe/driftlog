// Package lint provides rule-based validation of drift results,
// flagging results that violate defined policies (e.g. required tags,
// disallowed attribute values).
package lint

import (
	"fmt"
	"strings"

	"github.com/driftlog/internal/diff"
)

// Severity indicates how serious a lint violation is.
type Severity string

const (
	SeverityWarn  Severity = "warn"
	SeverityError Severity = "error"
)

// Violation describes a single rule breach found in a drift result.
type Violation struct {
	ResourceID string
	Rule       string
	Message    string
	Severity   Severity
}

// Rule is a function that inspects a DriftResult and returns any violations.
type Rule func(r diff.DriftResult) []Violation

// RequireTag returns a Rule that flags resources missing the given tag key
// in their cloud attributes.
func RequireTag(tag string) Rule {
	return func(r diff.DriftResult) []Violation {
		if r.Status == diff.StatusOnlyInState {
			return nil
		}
		for k := range r.CloudAttrs {
			if strings.EqualFold(k, tag) {
				return nil
			}
		}
		return []Violation{{
			ResourceID: r.ResourceID,
			Rule:       "require_tag",
			Message:    fmt.Sprintf("missing required tag %q", tag),
			Severity:   SeverityError,
		}}
	}
}

// DisallowAttributeValue returns a Rule that flags resources where
// the given attribute equals a disallowed value.
func DisallowAttributeValue(attr, value string) Rule {
	return func(r diff.DriftResult) []Violation {
		for k, v := range r.CloudAttrs {
			if strings.EqualFold(k, attr) && strings.EqualFold(v, value) {
				return []Violation{{
					ResourceID: r.ResourceID,
					Rule:       "disallow_attribute_value",
					Message:    fmt.Sprintf("attribute %q must not be %q", attr, value),
					Severity:   SeverityWarn,
				}}
			}
		}
		return nil
	}
}

// Run applies all rules to every result and returns the collected violations.
func Run(results []diff.DriftResult, rules []Rule) []Violation {
	var all []Violation
	for _, r := range results {
		for _, rule := range rules {
			all = append(all, rule(r)...)
		}
	}
	return all
}
