package ranking_test

import (
	"testing"

	"github.com/yourorg/driftlog/internal/diff"
	"github.com/yourorg/driftlog/internal/ranking"
)

func makeResult(id, rtype string, status diff.Status, changes int) diff.Result {
	ch := make([]diff.Change, changes)
	for i := range ch {
		ch[i] = diff.Change{Attribute: "attr", StateValue: "old", CloudValue: "new"}
	}
	return diff.Result{
		ResourceID:   id,
		ResourceType: rtype,
		Status:       status,
		Changes:      ch,
	}
}

func TestRank_SortsByScoreDescending(t *testing.T) {
	results := []diff.Result{
		makeResult("a", "aws_s3_bucket", diff.StatusClean, 0),
		makeResult("b", "aws_instance", diff.StatusDrifted, 2),
		makeResult("c", "aws_instance", diff.StatusOnlyInCloud, 0),
		makeResult("d", "aws_instance", diff.StatusOnlyInState, 0),
	}
	entries := ranking.Rank(results)
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}
	if entries[0].ResourceID != "c" {
		t.Errorf("expected first to be OnlyInCloud (c), got %s", entries[0].ResourceID)
	}
	if entries[1].ResourceID != "d" {
		t.Errorf("expected second to be OnlyInState (d), got %s", entries[1].ResourceID)
	}
	if entries[2].ResourceID != "b" {
		t.Errorf("expected third to be Drifted (b), got %s", entries[2].ResourceID)
	}
	if entries[3].ResourceID != "a" {
		t.Errorf("expected last to be Clean (a), got %s", entries[3].ResourceID)
	}
}

func TestRank_EmptyResults(t *testing.T) {
	entries := ranking.Rank(nil)
	if len(entries) != 0 {
		t.Errorf("expected empty, got %d", len(entries))
	}
}

func TestTopN_LimitsResults(t *testing.T) {
	results := []diff.Result{
		makeResult("a", "aws_instance", diff.StatusOnlyInCloud, 0),
		makeResult("b", "aws_instance", diff.StatusOnlyInState, 0),
		makeResult("c", "aws_instance", diff.StatusDrifted, 1),
	}
	top := ranking.TopN(results, 2)
	if len(top) != 2 {
		t.Errorf("expected 2, got %d", len(top))
	}
}

func TestTopN_NGreaterThanLen(t *testing.T) {
	results := []diff.Result{
		makeResult("a", "aws_instance", diff.StatusDrifted, 1),
	}
	top := ranking.TopN(results, 10)
	if len(top) != 1 {
		t.Errorf("expected 1, got %d", len(top))
	}
}

func TestRank_TieBreakByID(t *testing.T) {
	results := []diff.Result{
		makeResult("z-res", "aws_instance", diff.StatusOnlyInState, 0),
		makeResult("a-res", "aws_instance", diff.StatusOnlyInState, 0),
	}
	entries := ranking.Rank(results)
	if entries[0].ResourceID != "a-res" {
		t.Errorf("expected a-res first on tie-break, got %s", entries[0].ResourceID)
	}
}
