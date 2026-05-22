package summary_test

import (
	"strings"
	"testing"

	"github.com/driftlog/internal/diff"
	"github.com/driftlog/internal/summary"
)

func makeResults(statuses ...diff.DriftStatus) []diff.DriftResult {
	out := make([]diff.DriftResult, len(statuses))
	for i, s := range statuses {
		out[i] = diff.DriftResult{
			ResourceID:   fmt.Sprintf("res-%d", i),
			ResourceType: "aws_instance",
			Status:       s,
		}
	}
	return out
}

func TestBuild_EmptyResults(t *testing.T) {
	s := summary.Build(nil)
	if s.TotalResources != 0 {
		t.Fatalf("expected 0 total, got %d", s.TotalResources)
	}
	if s.DriftRate != 0 {
		t.Fatalf("expected 0 drift rate, got %f", s.DriftRate)
	}
}

func TestBuild_AllClean(t *testing.T) {
	results := makeResults(diff.StatusClean, diff.StatusClean, diff.StatusClean)
	s := summary.Build(results)
	if s.CleanCount != 3 {
		t.Fatalf("expected CleanCount=3, got %d", s.CleanCount)
	}
	if s.HasDrift() {
		t.Fatal("expected no drift")
	}
	if s.DriftRate != 0 {
		t.Fatalf("expected 0%% drift rate, got %.1f", s.DriftRate)
	}
}

func TestBuild_MixedStatuses(t *testing.T) {
	results := makeResults(
		diff.StatusClean,
		diff.StatusDrifted,
		diff.StatusMissing,
		diff.StatusUnmanaged,
	)
	s := summary.Build(results)
	if s.TotalResources != 4 {
		t.Fatalf("expected 4 total, got %d", s.TotalResources)
	}
	if s.DriftedCount != 1 || s.MissingCount != 1 || s.UnmanagedCount != 1 || s.CleanCount != 1 {
		t.Fatalf("unexpected counts: %+v", s)
	}
	if !s.HasDrift() {
		t.Fatal("expected drift to be detected")
	}
	if s.DriftRate != 75.0 {
		t.Fatalf("expected 75.0%% drift rate, got %.1f", s.DriftRate)
	}
}

func TestLines_ContainsExpectedFields(t *testing.T) {
	results := makeResults(diff.StatusClean, diff.StatusDrifted)
	s := summary.Build(results)
	lines := s.Lines()
	if len(lines) != 6 {
		t.Fatalf("expected 6 lines, got %d", len(lines))
	}
	for _, keyword := range []string{"Total", "Clean", "Drifted", "Missing", "Unmanaged", "Drift rate"} {
		found := false
		for _, l := range lines {
			if strings.Contains(l, keyword) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected line containing %q", keyword)
		}
	}
}
