package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/your-org/driftlog/internal/audit"
)

// runAudit prints the audit log found at the given path in a human-readable table.
func runAudit(path string) error {
	entries, err := audit.ReadAll(path)
	if err != nil {
		return fmt.Errorf("audit: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("No audit entries found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tRUN ID\tREGION\tTOTAL\tDRIFTED\tCLEAN\tMISSING\tUNMANAGED")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\n",
			e.Timestamp.Format("2006-01-02T15:04:05Z"),
			e.RunID,
			e.Region,
			e.TotalCount,
			e.DriftCount,
			e.CleanCount,
			e.MissingCount,
			e.UnmanagedCount,
		)
	}
	return w.Flush()
}
