package schedule_test

import (
	"testing"
	"time"

	"github.com/yourorg/driftlog/internal/schedule"
)

func TestParse_NamedIntervals(t *testing.T) {
	cases := []struct {
		input    string
		wantDur  time.Duration
		wantName string
	}{
		{"hourly", time.Hour, "hourly"},
		{"daily", 24 * time.Hour, "daily"},
		{"weekly", 7 * 24 * time.Hour, "weekly"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := schedule.Parse(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Duration != tc.wantDur {
				t.Errorf("duration: got %v, want %v", got.Duration, tc.wantDur)
			}
			if got.Name != tc.wantName {
				t.Errorf("name: got %q, want %q", got.Name, tc.wantName)
			}
		})
	}
}

func TestParse_CustomDuration(t *testing.T) {
	got, err := schedule.Parse("30m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Duration != 30*time.Minute {
		t.Errorf("got %v, want 30m", got.Duration)
	}
}

func TestParse_Invalid(t *testing.T) {
	if _, err := schedule.Parse("never"); err == nil {
		t.Error("expected error for unknown interval")
	}
	if _, err := schedule.Parse("-1h"); err == nil {
		t.Error("expected error for negative duration")
	}
}

func TestNextRun(t *testing.T) {
	last := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	next := schedule.NextRun(last, schedule.IntervalHourly)
	want := last.Add(time.Hour)
	if !next.Equal(want) {
		t.Errorf("NextRun: got %v, want %v", next, want)
	}
}

func TestIsDue(t *testing.T) {
	last := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	if schedule.IsDue(last, schedule.IntervalHourly, last.Add(30*time.Minute)) {
		t.Error("should not be due before interval elapses")
	}
	if !schedule.IsDue(last, schedule.IntervalHourly, last.Add(time.Hour)) {
		t.Error("should be due exactly at interval boundary")
	}
	if !schedule.IsDue(last, schedule.IntervalHourly, last.Add(2*time.Hour)) {
		t.Error("should be due after interval elapses")
	}
}
