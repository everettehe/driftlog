package score_test

import (
	"testing"

	"github.com/example/driftlog/internal/diff"
	"github.com/example/driftlog/internal/score"
)

func makeResult(status diff.Status) diff.Result {
	return diff.Result{ResourceID: "res-1", ResourceType: "aws_instance", Status: status}
}

func TestCompute_EmptyResults(t *testing.T) {
	r := score.Compute(nil)
	if r.Score != 100 {
		t.Errorf("expected 100, got %d", r.Score)
	}
	if r.Grade != "A" {
		t.Errorf("expected grade A, got %s", r.Grade)
	}
}

func TestCompute_AllClean(t *testing.T) {
	results := []diff.Result{
		makeResult(diff.StatusMatch),
		makeResult(diff.StatusMatch),
	}
	r := score.Compute(results)
	if r.Score != 100 {
		t.Errorf("expected 100, got %d", r.Score)
	}
	if r.Grade != "A" {
		t.Errorf("expected A, got %s", r.Grade)
	}
}

func TestCompute_AllMissing(t *testing.T) {
	results := []diff.Result{
		makeResult(diff.StatusOnlyInState),
		makeResult(diff.StatusOnlyInState),
	}
	r := score.Compute(results)
	if r.Score != 0 {
		t.Errorf("expected 0, got %d", r.Score)
	}
	if r.Grade != "F" {
		t.Errorf("expected F, got %s", r.Grade)
	}
	if r.Missing != 2 {
		t.Errorf("expected 2 missing, got %d", r.Missing)
	}
}

func TestCompute_MixedDrift(t *testing.T) {
	results := []diff.Result{
		makeResult(diff.StatusMatch),
		makeResult(diff.StatusDrifted),
		makeResult(diff.StatusOnlyInState),
		makeResult(diff.StatusOnlyInCloud),
	}
	r := score.Compute(results)
	// penalty = 1+2+1 = 4; maxPenalty = 4*2 = 8; score = round(100*(1-4/8)) = 50
	if r.Score != 50 {
		t.Errorf("expected 50, got %d", r.Score)
	}
	if r.Grade != "D" {
		t.Errorf("expected D, got %s", r.Grade)
	}
	if r.Drifted != 1 || r.Missing != 1 || r.Untracked != 1 {
		t.Errorf("unexpected counts: drifted=%d missing=%d untracked=%d",
			r.Drifted, r.Missing, r.Untracked)
	}
}

func TestLines_ContainsScore(t *testing.T) {
	r := score.Compute([]diff.Result{makeResult(diff.StatusMatch)})
	lines := r.Lines()
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	for _, l := range lines {
		if l == "" {
			t.Error("unexpected empty line")
		}
	}
}
