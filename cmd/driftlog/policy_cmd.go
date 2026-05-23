package main

import (
	"fmt"
	"os"

	"github.com/driftlog/internal/diff"
	"github.com/driftlog/internal/policy"
)

// runPolicy evaluates built-in and configured policy rules against drift results
// and prints a human-readable report. Exits with code 1 if any errors are found.
func runPolicy(results []diff.Result, extraRules []policy.Rule) error {
	rules := policy.DefaultRules()
	rules = append(rules, extraRules...)

	report := policy.Evaluate(results, rules)
	fmt.Println(policy.FormatReport(report))

	if report.HasErrors() {
		os.Exit(1)
	}
	return nil
}
