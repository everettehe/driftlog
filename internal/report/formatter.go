package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/user/driftlog/internal/diff"
)

// Format controls the output format of the report.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Summary holds aggregate counts from a drift report.
type Summary struct {
	Total    int
	Drifted  int
	Missing  int
	Orphaned int
}

// WriteText writes a human-readable drift report to w.
func WriteText(w io.Writer, results []diff.DriftResult) error {
	s := summarize(results)
	fmt.Fprintf(w, "Drift Report\n")
	fmt.Fprintf(w, "%s\n", strings.Repeat("=", 40))

	for _, r := range results {
		if r.Status == diff.StatusMatch {
			continue
		}
		fmt.Fprintf(w, "\n[%s] %s (%s)\n", statusLabel(r.Status), r.ResourceID, r.ResourceType)
		if len(r.Diffs) > 0 {
			for _, d := range r.Diffs {
				fmt.Fprintf(w, "  %s\n", d)
			}
		}
	}

	fmt.Fprintf(w, "\n%s\n", strings.Repeat("-", 40))
	fmt.Fprintf(w, "Summary: %d total | %d drifted | %d missing | %d orphaned\n",
		s.Total, s.Drifted, s.Missing, s.Orphaned)
	return nil
}

func summarize(results []diff.DriftResult) Summary {
	s := Summary{Total: len(results)}
	for _, r := range results {
		switch r.Status {
		case diff.StatusDrifted:
			s.Drifted++
		case diff.StatusMissing:
			s.Missing++
		case diff.StatusOrphaned:
			s.Orphaned++
		}
	}
	return s
}

func statusLabel(s diff.DriftStatus) string {
	switch s {
	case diff.StatusDrifted:
		return "DRIFT"
	case diff.StatusMissing:
		return "MISSING"
	case diff.StatusOrphaned:
		return "ORPHANED"
	default:
		return "OK"
	}
}
