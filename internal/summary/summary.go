// Package summary provides run summary aggregation across multiple drift scans.
package summary

import (
	"fmt"
	"time"

	"github.com/driftlog/internal/diff"
)

// RunSummary holds aggregated statistics for a single drift detection run.
type RunSummary struct {
	RunAt          time.Time `json:"run_at"`
	TotalResources int       `json:"total_resources"`
	DriftedCount   int       `json:"drifted_count"`
	MissingCount   int       `json:"missing_count"`
	UnmanagedCount int       `json:"unmanaged_count"`
	CleanCount     int       `json:"clean_count"`
	DriftRate      float64   `json:"drift_rate_pct"`
}

// Build constructs a RunSummary from a slice of DriftResult values.
func Build(results []diff.DriftResult) RunSummary {
	s := RunSummary{
		RunAt:          time.Now().UTC(),
		TotalResources: len(results),
	}

	for _, r := range results {
		switch r.Status {
		case diff.StatusDrifted:
			s.DriftedCount++
		case diff.StatusMissing:
			s.MissingCount++
		case diff.StatusUnmanaged:
			s.UnmanagedCount++
		case diff.StatusClean:
			s.CleanCount++
		}
	}

	if s.TotalResources > 0 {
		drifted := s.DriftedCount + s.MissingCount + s.UnmanagedCount
		s.DriftRate = float64(drifted) / float64(s.TotalResources) * 100.0
	}

	return s
}

// HasDrift returns true if any resource is not clean.
func (s RunSummary) HasDrift() bool {
	return s.DriftedCount > 0 || s.MissingCount > 0 || s.UnmanagedCount > 0
}

// Lines returns a slice of human-readable summary lines.
func (s RunSummary) Lines() []string {
	return []string{
		fmt.Sprintf("Total resources : %d", s.TotalResources),
		fmt.Sprintf("Clean           : %d", s.CleanCount),
		fmt.Sprintf("Drifted         : %d", s.DriftedCount),
		fmt.Sprintf("Missing         : %d", s.MissingCount),
		fmt.Sprintf("Unmanaged       : %d", s.UnmanagedCount),
		fmt.Sprintf("Drift rate      : %.1f%%", s.DriftRate),
	}
}
