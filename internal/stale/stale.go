// Package stale identifies resources that have not changed between two
// consecutive drift scans, flagging them as potentially stale or ignored.
package stale

import (
	"fmt"
	"time"

	"github.com/driftlog/internal/diff"
)

// Entry describes a resource that has appeared unchanged across multiple runs.
type Entry struct {
	ResourceID string
	ResourceType string
	Status diff.Status
	FirstSeenAt time.Time
	LastSeenAt time.Time
	Occurrences int
}

// Report holds all stale entries detected.
type Report struct {
	Entries []Entry
	GeneratedAt time.Time
}

// Detect compares two slices of DriftResult (previous and current) and returns
// a Report of resources whose drift status has not changed between the two runs.
func Detect(previous, current []diff.DriftResult, since time.Time) Report {
	prev := indexByID(previous)
	report := Report{GeneratedAt: time.Now()}

	for _, cur := range current {
		p, ok := prev[cur.ResourceID]
		if !ok {
			continue
		}
		if p.Status != cur.Status {
			continue
		}
		// Same resource, same status — consider it stale.
		entry := Entry{
			ResourceID: cur.ResourceID,
			ResourceType: cur.ResourceType,
			Status: cur.Status,
			FirstSeenAt: since,
			LastSeenAt: time.Now(),
			Occurrences: 2,
		}
		report.Entries = append(report.Entries, entry)
	}
	return report
}

// Lines returns a human-readable summary of the stale report.
func Lines(r Report) []string {
	if len(r.Entries) == 0 {
		return []string{"No stale resources detected."}
	}
	lines := []string{fmt.Sprintf("Stale resources (%d):", len(r.Entries))}
	for _, e := range r.Entries {
		lines = append(lines, fmt.Sprintf(
			"  [%s] %s (%s) — unchanged since %s",
			e.Status, e.ResourceID, e.ResourceType,
			e.FirstSeenAt.Format(time.RFC3339),
		))
	}
	return lines
}

func indexByID(results []diff.DriftResult) map[string]diff.DriftResult {
	m := make(map[string]diff.DriftResult, len(results))
	for _, r := range results {
		m[r.ResourceID] = r
	}
	return m
}
