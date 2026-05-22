package metrics

import (
	"strings"
	"testing"
	"time"

	"github.com/driftlog/internal/diff"
)

func makeResults() []diff.DriftResult {
	return []diff.DriftResult{
		{ResourceID: "i-1", Status: diff.StatusOK},
		{ResourceID: "i-2", Status: diff.StatusDrifted},
		{ResourceID: "i-3", Status: diff.StatusDrifted},
		{ResourceID: "i-4", Status: diff.StatusMissing},
		{ResourceID: "i-5", Status: diff.StatusExtra},
	}
}

func TestNew_SetsStartedAt(t *testing.T) {
	before := time.Now()
	m := New()
	after := time.Now()

	if m.StartedAt.Before(before) || m.StartedAt.After(after) {
		t.Errorf("StartedAt %v not in expected range", m.StartedAt)
	}
}

func TestRecord_CountsCorrectly(t *testing.T) {
	m := New()
	m.Record(makeResults())

	if m.TotalChecked != 5 {
		t.Errorf("TotalChecked: want 5, got %d", m.TotalChecked)
	}
	if m.Drifted != 2 {
		t.Errorf("Drifted: want 2, got %d", m.Drifted)
	}
	if m.Missing != 1 {
		t.Errorf("Missing: want 1, got %d", m.Missing)
	}
	if m.Extra != 1 {
		t.Errorf("Extra: want 1, got %d", m.Extra)
	}
}

func TestRecord_EmptyResults(t *testing.T) {
	m := New()
	m.Record([]diff.DriftResult{})

	if m.TotalChecked != 0 || m.Drifted != 0 {
		t.Errorf("expected all zeros for empty results")
	}
}

func TestFinish_SetsFinishedAt(t *testing.T) {
	m := New()
	m.Finish()

	if m.FinishedAt.IsZero() {
		t.Error("FinishedAt should not be zero after Finish()")
	}
}

func TestDuration_BeforeFinish(t *testing.T) {
	m := New()
	time.Sleep(5 * time.Millisecond)
	d := m.Duration()

	if d < 5*time.Millisecond {
		t.Errorf("Duration before Finish should be >= 5ms, got %v", d)
	}
}

func TestDuration_AfterFinish(t *testing.T) {
	m := New()
	time.Sleep(5 * time.Millisecond)
	m.Finish()
	d := m.Duration()

	if d < 5*time.Millisecond {
		t.Errorf("Duration after Finish should be >= 5ms, got %v", d)
	}
}

func TestSummary_ContainsKeyFields(t *testing.T) {
	m := New()
	m.Record(makeResults())
	m.Finish()

	s := m.Summary()
	for _, want := range []string{"checked=5", "drifted=2", "missing=1", "extra=1", "errors=0"} {
		if !strings.Contains(s, want) {
			t.Errorf("Summary missing %q in: %s", want, s)
		}
	}
}
