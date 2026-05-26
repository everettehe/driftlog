package rollup_test

import (
	"testing"
	"time"

	"github.com/you/driftlog/internal/diff"
	"github.com/you/driftlog/internal/rollup"
)

func makeEntry(day string, results []diff.Result) rollup.Entry {
	t, _ := time.Parse("2006-01-02", day)
	return rollup.Entry{At: t, Results: results}
}

func makeResult(status diff.Status) diff.Result {
	return diff.Result{ResourceID: "r", Status: status}
}

func TestByDay_EmptyEntries(t *testing.T) {
	r := rollup.ByDay(nil)
	if len(r.Periods) != 0 {
		t.Fatalf("expected 0 periods, got %d", len(r.Periods))
	}
}

func TestByDay_SingleDay(t *testing.T) {
	entries := []rollup.Entry{
		makeEntry("2024-03-01", []diff.Result{
			makeResult(diff.StatusDrifted),
			makeResult(diff.StatusMatch),
		}),
	}
	r := rollup.ByDay(entries)
	if len(r.Periods) != 1 {
		t.Fatalf("expected 1 period, got %d", len(r.Periods))
	}
	p := r.Periods[0]
	if p.Label != "2024-03-01" {
		t.Errorf("unexpected label %q", p.Label)
	}
	if p.Total != 2 || p.Drifted != 1 {
		t.Errorf("unexpected counts total=%d drifted=%d", p.Total, p.Drifted)
	}
}

func TestByDay_MultipleDays_SortedAscending(t *testing.T) {
	entries := []rollup.Entry{
		makeEntry("2024-03-03", []diff.Result{makeResult(diff.StatusOnlyInCloud)}),
		makeEntry("2024-03-01", []diff.Result{makeResult(diff.StatusOnlyInState)}),
		makeEntry("2024-03-02", []diff.Result{makeResult(diff.StatusDrifted)}),
	}
	r := rollup.ByDay(entries)
	if len(r.Periods) != 3 {
		t.Fatalf("expected 3 periods, got %d", len(r.Periods))
	}
	if r.Periods[0].Label != "2024-03-01" || r.Periods[2].Label != "2024-03-03" {
		t.Errorf("periods not sorted: %v", r.Periods)
	}
}

func TestByDay_StatusCounts(t *testing.T) {
	entries := []rollup.Entry{
		makeEntry("2024-04-01", []diff.Result{
			makeResult(diff.StatusDrifted),
			makeResult(diff.StatusDrifted),
			makeResult(diff.StatusOnlyInState),
			makeResult(diff.StatusOnlyInCloud),
			makeResult(diff.StatusMatch),
		}),
	}
	r := rollup.ByDay(entries)
	p := r.Periods[0]
	if p.Drifted != 2 {
		t.Errorf("expected 2 drifted, got %d", p.Drifted)
	}
	if p.Missing != 1 {
		t.Errorf("expected 1 missing, got %d", p.Missing)
	}
	if p.Unmanaged != 1 {
		t.Errorf("expected 1 unmanaged, got %d", p.Unmanaged)
	}
	if p.Total != 5 {
		t.Errorf("expected 5 total, got %d", p.Total)
	}
}

func TestByWeek_GroupsIntoWeeks(t *testing.T) {
	entries := []rollup.Entry{
		makeEntry("2024-03-04", []diff.Result{makeResult(diff.StatusDrifted)}), // W10
		makeEntry("2024-03-07", []diff.Result{makeResult(diff.StatusMatch)}),   // W10
		makeEntry("2024-03-11", []diff.Result{makeResult(diff.StatusDrifted)}), // W11
	}
	r := rollup.ByWeek(entries)
	if len(r.Periods) != 2 {
		t.Fatalf("expected 2 weekly periods, got %d", len(r.Periods))
	}
	if r.Periods[0].Total != 2 {
		t.Errorf("W10 should have 2 results, got %d", r.Periods[0].Total)
	}
}
