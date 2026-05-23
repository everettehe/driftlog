package policy

import (
	"fmt"
	"strings"

	"github.com/driftlog/internal/diff"
)

// Rule defines a single policy rule applied to drift results.
type Rule struct {
	Name        string
	Description string
	Severity    string // "error", "warn", "info"
	Check       func(r diff.Result) *Violation
}

// Violation represents a policy rule breach.
type Violation struct {
	Rule       string
	Severity   string
	ResourceID string
	Message    string
}

// Report holds the outcome of evaluating all rules.
type Report struct {
	Violations []Violation
	Passed     int
	Failed     int
}

// HasErrors returns true if any violation has severity "error".
func (r *Report) HasErrors() bool {
	for _, v := range r.Violations {
		if v.Severity == "error" {
			return true
		}
	}
	return false
}

// Evaluate runs all rules against the provided results and returns a Report.
func Evaluate(results []diff.Result, rules []Rule) Report {
	report := Report{}
	for _, result := range results {
		for _, rule := range rules {
			if v := rule.Check(result); v != nil {
				v.Rule = rule.Name
				v.Severity = rule.Severity
				v.ResourceID = result.ResourceID
				report.Violations = append(report.Violations, *v)
				report.Failed++
			} else {
				report.Passed++
			}
		}
	}
	return report
}

// FormatReport returns a human-readable string of all violations.
func FormatReport(r Report) string {
	if len(r.Violations) == 0 {
		return fmt.Sprintf("Policy check passed. %d checks OK.", r.Passed)
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Policy violations: %d (passed: %d)\n", r.Failed, r.Passed))
	for _, v := range r.Violations {
		sb.WriteString(fmt.Sprintf("  [%s] %s — %s: %s\n",
			strings.ToUpper(v.Severity), v.ResourceID, v.Rule, v.Message))
	}
	return strings.TrimRight(sb.String(), "\n")
}
