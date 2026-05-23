// Package trend analyzes drift history to surface patterns and trends over time.
package trend

import (
	"fmt"
	"time"

	"github.com/user/driftlog/internal/history"
)

// Direction indicates whether drift is increasing, decreasing, or stable.
type Direction string

const (
	DirectionUp     Direction = "increasing"
	DirectionDown   Direction = "decreasing"
	DirectionStable Direction = "stable"
)

// Summary holds a trend analysis over a set of historical runs.
type Summary struct {
	TotalRuns    int
	AvgDrift     float64
	PeakDrift    int
	PeakAt       time.Time
	LatestDrift  int
	Direction    Direction
	WindowStart  time.Time
	WindowEnd    time.Time
}

// Analyze computes a trend summary from the given history entries.
// It considers only entries within the last windowDays days.
func Analyze(entries []history.Entry, windowDays int) Summary {
	if len(entries) == 0 {
		return Summary{}
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -windowDays)
	var windowed []history.Entry
	for _, e := range entries {
		if e.RunAt.After(cutoff) {
			windowed = append(windowed, e)
		}
	}
	if len(windowed) == 0 {
		return Summary{}
	}

	var total int
	peak := windowed[0].DriftCount
	peakAt := windowed[0].RunAt

	for _, e := range windowed {
		total += e.DriftCount
		if e.DriftCount > peak {
			peak = e.DriftCount
			peakAt = e.RunAt
		}
	}

	avg := float64(total) / float64(len(windowed))
	latest := windowed[len(windowed)-1].DriftCount
	dir := direction(windowed)

	return Summary{
		TotalRuns:   len(windowed),
		AvgDrift:    avg,
		PeakDrift:   peak,
		PeakAt:      peakAt,
		LatestDrift: latest,
		Direction:   dir,
		WindowStart: windowed[0].RunAt,
		WindowEnd:   windowed[len(windowed)-1].RunAt,
	}
}

// Lines returns a human-readable slice of strings describing the trend.
func Lines(s Summary) []string {
	if s.TotalRuns == 0 {
		return []string{"No trend data available."}
	}
	return []string{
		fmt.Sprintf("Runs analysed : %d", s.TotalRuns),
		fmt.Sprintf("Window        : %s → %s",
			s.WindowStart.Format("2006-01-02"),
			s.WindowEnd.Format("2006-01-02")),
		fmt.Sprintf("Avg drift     : %.1f resources/run", s.AvgDrift),
		fmt.Sprintf("Peak drift    : %d (on %s)", s.PeakDrift, s.PeakAt.Format("2006-01-02")),
		fmt.Sprintf("Latest drift  : %d", s.LatestDrift),
		fmt.Sprintf("Trend         : %s", s.Direction),
	}
}

func direction(entries []history.Entry) Direction {
	if len(entries) < 2 {
		return DirectionStable
	}
	first := entries[0].DriftCount
	last := entries[len(entries)-1].DriftCount
	switch {
	case last > first:
		return DirectionUp
	case last < first:
		return DirectionDown
	default:
		return DirectionStable
	}
}
