package main

import (
	"fmt"
	"os"

	"github.com/driftlog/internal/diff"
	"github.com/driftlog/internal/remediate"
)

// runRemediate prints remediation suggestions for the given drift results.
// It is called after a normal drift detection run when --remediate is set.
func runRemediate(results []diff.Result, outputPath string) error {
	actions := remediate.Suggest(results)
	formatted := remediate.Format(actions)

	if outputPath == "" || outputPath == "-" {
		fmt.Print(formatted)
		return nil
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("remediate: create output file: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprint(f, formatted)
	if err != nil {
		return fmt.Errorf("remediate: write output: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Remediation suggestions written to %s\n", outputPath)
	return nil
}
