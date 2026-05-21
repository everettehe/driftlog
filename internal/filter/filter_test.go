package filter_test

import (
	"testing"

	"github.com/your-org/driftlog/internal/diff"
	"github.com/your-org/driftlog/internal/filter"
)

func makeResult(id, rtype string, status diff.DriftStatus) diff.DriftResult {
	return diff.DriftResult{
		ResourceID:   id,
		ResourceType: rtype,
		Status:       status,
	}
}

var sampleResults = []diff.DriftResult{
	makeResult("i-001", "aws_instance", diff.StatusDrifted),
	makeResult("i-002", "aws_instance", diff.StatusMatch),
	makeResult("bucket-a", "aws_s3_bucket", diff.StatusDrifted),
	makeResult("sg-01", "aws_security_group", diff.StatusOnlyInState),
}

func TestApply_NoOptions(t *testing.T) {
	out := filter.Apply(sampleResults, filter.Options{})
	if len(out) != len(sampleResults) {
		t.Fatalf("expected %d results, got %d", len(sampleResults), len(out))
	}
}

func TestApply_OnlyDrifted(t *testing.T) {
	out := filter.Apply(sampleResults, filter.Options{OnlyDrifted: true})
	for _, r := range out {
		if r.Status == diff.StatusMatch {
			t.Errorf("unexpected match result in OnlyDrifted output: %s", r.ResourceID)
		}
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 drifted results, got %d", len(out))
	}
}

func TestApply_FilterByType(t *testing.T) {
	out := filter.Apply(sampleResults, filter.Options{ResourceTypes: []string{"aws_instance"}})
	if len(out) != 2 {
		t.Fatalf("expected 2 aws_instance results, got %d", len(out))
	}
}

func TestApply_FilterByType_CaseInsensitive(t *testing.T) {
	out := filter.Apply(sampleResults, filter.Options{ResourceTypes: []string{"AWS_S3_BUCKET"}})
	if len(out) != 1 || out[0].ResourceID != "bucket-a" {
		t.Fatalf("expected bucket-a, got %+v", out)
	}
}

func TestApply_ExcludeIDs(t *testing.T) {
	out := filter.Apply(sampleResults, filter.Options{ExcludeIDs: []string{"i-001", "sg-01"}})
	if len(out) != 2 {
		t.Fatalf("expected 2 results after exclusion, got %d", len(out))
	}
	for _, r := range out {
		if r.ResourceID == "i-001" || r.ResourceID == "sg-01" {
			t.Errorf("excluded resource %s still present", r.ResourceID)
		}
	}
}

func TestApply_CombinedOptions(t *testing.T) {
	opts := filter.Options{
		OnlyDrifted:   true,
		ResourceTypes: []string{"aws_instance"},
	}
	out := filter.Apply(sampleResults, opts)
	if len(out) != 1 || out[0].ResourceID != "i-001" {
		t.Fatalf("expected only i-001, got %+v", out)
	}
}
