package schedule

import (
	"fmt"
	"time"
)

// Interval represents a named drift-check schedule.
type Interval struct {
	Name     string
	Duration time.Duration
}

// Predefined schedule intervals.
var (
	IntervalHourly  = Interval{Name: "hourly", Duration: time.Hour}
	IntervalDaily   = Interval{Name: "daily", Duration: 24 * time.Hour}
	IntervalWeekly  = Interval{Name: "weekly", Duration: 7 * 24 * time.Hour}
)

// Parse converts a string like "hourly", "daily", "weekly", or a Go duration
// string (e.g. "30m", "6h") into an Interval.
func Parse(s string) (Interval, error) {
	switch s {
	case "hourly":
		return IntervalHourly, nil
	case "daily":
		return IntervalDaily, nil
	case "weekly":
		return IntervalWeekly, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return Interval{}, fmt.Errorf("schedule: unrecognised interval %q", s)
	}
	if d <= 0 {
		return Interval{}, fmt.Errorf("schedule: interval must be positive, got %q", s)
	}
	return Interval{Name: s, Duration: d}, nil
}

// NextRun returns the next time a run should occur given the last run time and
// the desired interval.
func NextRun(last time.Time, interval Interval) time.Time {
	return last.Add(interval.Duration)
}

// IsDue reports whether a new drift check should be triggered.
// It returns true when the current time is at or after the next scheduled run.
func IsDue(last time.Time, interval Interval, now time.Time) bool {
	return !now.Before(NextRun(last, interval))
}
