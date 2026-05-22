package tags_test

import (
	"strings"
	"testing"

	"github.com/yourusername/driftlog/internal/tags"
)

func TestCompare_NoDiff(t *testing.T) {
	state := map[string]string{"env": "prod", "team": "platform"}
	live := map[string]string{"env": "prod", "team": "platform"}
	diffs := tags.Compare(state, live)
	if len(diffs) != 0 {
		t.Fatalf("expected no diffs, got %d", len(diffs))
	}
}

func TestCompare_ChangedValue(t *testing.T) {
	state := map[string]string{"env": "staging"}
	live := map[string]string{"env": "prod"}
	diffs := tags.Compare(state, live)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	d := diffs[0]
	if d.Key != "env" || d.Status != "changed" || d.Expected != "staging" || d.Actual != "prod" {
		t.Errorf("unexpected diff: %+v", d)
	}
}

func TestCompare_RemovedTag(t *testing.T) {
	state := map[string]string{"owner": "alice"}
	live := map[string]string{}
	diffs := tags.Compare(state, live)
	if len(diffs) != 1 || diffs[0].Status != "removed" {
		t.Fatalf("expected 1 removed diff, got %+v", diffs)
	}
}

func TestCompare_AddedTag(t *testing.T) {
	state := map[string]string{}
	live := map[string]string{"cost-center": "42"}
	diffs := tags.Compare(state, live)
	if len(diffs) != 1 || diffs[0].Status != "added" {
		t.Fatalf("expected 1 added diff, got %+v", diffs)
	}
}

func TestCompare_SortedByKey(t *testing.T) {
	state := map[string]string{"z": "1", "a": "1"}
	live := map[string]string{"z": "2", "a": "2"}
	diffs := tags.Compare(state, live)
	if len(diffs) != 2 {
		t.Fatalf("expected 2 diffs")
	}
	if diffs[0].Key != "a" || diffs[1].Key != "z" {
		t.Errorf("diffs not sorted: %v, %v", diffs[0].Key, diffs[1].Key)
	}
}

func TestFormat_Empty(t *testing.T) {
	out := tags.Format(nil)
	if out != "" {
		t.Errorf("expected empty string, got %q", out)
	}
}

func TestFormat_ContainsSymbols(t *testing.T) {
	diffs := []tags.Diff{
		{Key: "env", Expected: "staging", Actual: "prod", Status: "changed"},
		{Key: "owner", Expected: "alice", Actual: "", Status: "removed"},
		{Key: "new-tag", Expected: "", Actual: "yes", Status: "added"},
	}
	out := tags.Format(diffs)
	if !strings.Contains(out, "~") {
		t.Error("expected '~' for changed tag")
	}
	if !strings.Contains(out, "-") {
		t.Error("expected '-' for removed tag")
	}
	if !strings.Contains(out, "+") {
		t.Error("expected '+' for added tag")
	}
}

func TestHasDrift_True(t *testing.T) {
	if !tags.HasDrift(map[string]string{"k": "a"}, map[string]string{"k": "b"}) {
		t.Error("expected drift")
	}
}

func TestHasDrift_False(t *testing.T) {
	if tags.HasDrift(map[string]string{"k": "v"}, map[string]string{"k": "v"}) {
		t.Error("expected no drift")
	}
}
