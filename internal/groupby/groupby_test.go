package groupby_test

import (
	"testing"

	"github.com/yourorg/driftlog/internal/diff"
	"github.com/yourorg/driftlog/internal/groupby"
)

func makeResult(id, rtype string, status diff.Status, attrs map[string]string) diff.Result {
	return diff.Result{
		ResourceID:   id,
		ResourceType: rtype,
		Status:       status,
		Attributes:   attrs,
	}
}

func TestByType_GroupsCorrectly(t *testing.T) {
	results := []diff.Result{
		makeResult("i-1", "aws_instance", diff.StatusDrifted, nil),
		makeResult("b-1", "aws_s3_bucket", diff.StatusClean, nil),
		makeResult("i-2", "aws_instance", diff.StatusClean, nil),
	}

	groups := groupby.ByType(results)

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Label != "aws_instance" {
		t.Errorf("expected first group aws_instance, got %s", groups[0].Label)
	}
	if len(groups[0].Results) != 2 {
		t.Errorf("expected 2 results in aws_instance group, got %d", len(groups[0].Results))
	}
}

func TestByStatus_GroupsCorrectly(t *testing.T) {
	results := []diff.Result{
		makeResult("i-1", "aws_instance", diff.StatusDrifted, nil),
		makeResult("i-2", "aws_instance", diff.StatusClean, nil),
		makeResult("i-3", "aws_instance", diff.StatusDrifted, nil),
	}

	groups := groupby.ByStatus(results)

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	for _, g := range groups {
		if g.Label == string(diff.StatusDrifted) && len(g.Results) != 2 {
			t.Errorf("expected 2 drifted results, got %d", len(g.Results))
		}
	}
}

func TestByTag_GroupsByTagValue(t *testing.T) {
	results := []diff.Result{
		makeResult("i-1", "aws_instance", diff.StatusClean, map[string]string{"env": "prod"}),
		makeResult("i-2", "aws_instance", diff.StatusClean, map[string]string{"env": "staging"}),
		makeResult("i-3", "aws_instance", diff.StatusClean, map[string]string{"env": "prod"}),
		makeResult("i-4", "aws_instance", diff.StatusClean, nil),
	}

	groups := groupby.ByTag(results, "env")

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}

	labels := map[string]int{}
	for _, g := range groups {
		labels[g.Label] = len(g.Results)
	}

	if labels["env=prod"] != 2 {
		t.Errorf("expected 2 prod results, got %d", labels["env=prod"])
	}
	if labels["<untagged>"] != 1 {
		t.Errorf("expected 1 untagged result, got %d", labels["<untagged>"])
	}
}

func TestByType_Empty(t *testing.T) {
	groups := groupby.ByType(nil)
	if len(groups) != 0 {
		t.Errorf("expected 0 groups for nil input, got %d", len(groups))
	}
}
