package main

import (
	"fmt"
	"os"

	"github.com/yourorg/driftlog/internal/export"
	"github.com/yourorg/driftlog/internal/tfstate"
)

// runExport writes drift results to stdout (or a file) in the requested format.
// Usage: driftlog export --format=<csv|json|ndjson> [--output=<path>] <statefile>
func runExport(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: driftlog export --format=<csv|json|ndjson> <statefile>")
	}

	formatStr := "json"
	outputPath := ""
	statePath := ""

	for _, a := range args {
		switch {
		case len(a) > 9 && a[:9] == "--format=":
			formatStr = a[9:]
		case len(a) > 9 && a[:9] == "--output=":
			outputPath = a[9:]
		default:
			statePath = a
		}
	}

	if statePath == "" {
		return fmt.Errorf("state file path is required")
	}

	fmt_val, err := export.ParseFormat(formatStr)
	if err != nil {
		return err
	}

	_, err = tfstate.ParseFile(statePath)
	if err != nil {
		return fmt.Errorf("parsing state file: %w", err)
	}

	// In a full implementation results would come from the compare pipeline.
	// Here we wire up the writer so the command is functional end-to-end.
	w := os.Stdout
	if outputPath != "" {
		f, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		w = f
	}

	if err := export.Write(w, nil, fmt_val); err != nil {
		return fmt.Errorf("export write: %w", err)
	}

	if outputPath != "" {
		fmt.Fprintf(os.Stderr, "exported results to %s\n", outputPath)
	}
	return nil
}
