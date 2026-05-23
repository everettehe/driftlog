package trend_test

import (
	"testing"
	"time"

	"github.com/user/driftlog/internal/history"
	"github.com/user/driftlog/internal/trend"
)

func makeEntry(daysAgo int, driftCount int) history.Entry {
	return history.Entry{
		RunAt:      time.Now().UTC().AddDate(0, 0, -daysAgo),
		DriftCount: driftCount,
	}
}

func TestAnalyze_EmptyEntries(t *testing.T) {
	s := trend.Analyze(nil, 7)
	if s.TotalRuns != 0 {
		t.Errorf("expected 0 runs, got %d", s.TotalRuns)
	}
}

func TestAnalyze_AllOutsideWindow(t *testing.T) {
	entries := []history.Entry{makeEntry(30, 5), makeEntry(20, 3)}
	s := trend.Analyze(entries, 7)
	if s.TotalRuns != 0 {
		t.Errorf("expected 0 runs within window, got %d", s.TotalRuns)
	}
}

func TestAnalyze_BasicCounts(t *testing.T) {
	entries := []history.Entry{
		makeEntry(6, 2),
		makeEntry(4, 4),
		makeEntry(2, 6),
	}
	s := trend.Analyze(entries, 7)
	if s.TotalRuns != 3 {
		t.Errorf("expected 3 runs, got %d", s.TotalRuns)
	}
	if s.PeakDrift != 6 {
		t.Errorf("expected peak 6, got %d", s.PeakDrift)
	}
	if s.LatestDrift != 6 {
		t.Errorf("expected latest 6, got %d", s.LatestDrift)
	}
	expectedAvg := 4.0
	if s.AvgDrift != expectedAvg {
		t.Errorf("expected avg %.1f, got %.1f", expectedAvg, s.AvgDrift)
	}
}

func TestAnalyze_DirectionIncreasing(t *testing.T) {
	entries := []history.Entry{makeEntry(5, 1), makeEntry(3, 3), makeEntry(1, 7)}
	s := trend.Analyze(entries, 7)
	if s.Direction != trend.DirectionUp {
		t.Errorf("expected increasing, got %s", s.Direction)
	}
}

func TestAnalyze_DirectionDecreasing(t *testing.T) {
	entries := []history.Entry{makeEntry(5, 8), makeEntry(3, 4), makeEntry(1, 1)}
	s := trend.Analyze(entries, 7)
	if s.Direction != trend.DirectionDown {
		t.Errorf("expected decreasing, got %s", s.Direction)
	}
}

func TestAnalyze_DirectionStable(t *testing.T) {
	entries := []history.Entry{makeEntry(5, 3), makeEntry(3, 5), makeEntry(1, 3)}
	s := trend.Analyze(entries, 7)
	if s.Direction != trend.DirectionStable {
		t.Errorf("expected stable, got %s", s.Direction)
	}
}

func TestLines_EmptySummary(t *testing.T) {
	lines := trend.Lines(trend.Summary{})
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0] != "No trend data available." {
		t.Errorf("unexpected message: %s", lines[0])
	}
}

func TestLines_PopulatedSummary(t *testing.T) {
	entries := []history.Entry{makeEntry(6, 2), makeEntry(2, 6)}
	s := trend.Analyze(entries, 7)
	lines := trend.Lines(s)
	if len(lines) != 6 {
		t.Errorf("expected 6 lines, got %d", len(lines))
	}
}
