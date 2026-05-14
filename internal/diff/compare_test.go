package diff

import (
	"strings"
	"testing"
)

func makeResource(id, rtype string, attrs map[string]string) Resource {
	return Resource{ID: id, Type: rtype, Attributes: attrs}
}

func TestCompare_NoDrift(t *testing.T) {
	state := []Resource{makeResource("i-123", "aws_instance", map[string]string{"instance_type": "t2.micro"})}
	cloud := []Resource{makeResource("i-123", "aws_instance", map[string]string{"instance_type": "t2.micro"})}

	diffs := Compare(state, cloud)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs, got %d", len(diffs))
	}
}

func TestCompare_AttributeDrift(t *testing.T) {
	state := []Resource{makeResource("i-123", "aws_instance", map[string]string{"instance_type": "t2.micro"})}
	cloud := []Resource{makeResource("i-123", "aws_instance", map[string]string{"instance_type": "t3.medium"})}

	diffs := Compare(state, cloud)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if len(diffs[0].Changes) != 1 {
		t.Fatalf("expected 1 attribute change, got %d", len(diffs[0].Changes))
	}
	c := diffs[0].Changes[0]
	if c.Attribute != "instance_type" || c.StateValue != "t2.micro" || c.CloudValue != "t3.medium" {
		t.Errorf("unexpected change: %+v", c)
	}
}

func TestCompare_OnlyInState(t *testing.T) {
	state := []Resource{makeResource("i-999", "aws_instance", map[string]string{})}
	cloud := []Resource{}

	diffs := Compare(state, cloud)
	if len(diffs) != 1 || !diffs[0].OnlyInState {
		t.Errorf("expected OnlyInState diff, got %+v", diffs)
	}
}

func TestCompare_OnlyInCloud(t *testing.T) {
	state := []Resource{}
	cloud := []Resource{makeResource("i-888", "aws_instance", map[string]string{})}

	diffs := Compare(state, cloud)
	if len(diffs) != 1 || !diffs[0].OnlyInCloud {
		t.Errorf("expected OnlyInCloud diff, got %+v", diffs)
	}
}

func TestFormatDiff_NoDrift(t *testing.T) {
	out := FormatDiff(nil)
	if out != "No drift detected." {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestFormatDiff_WithDrift(t *testing.T) {
	diffs := []ResourceDiff{
		{
			ResourceID:   "i-123",
			ResourceType: "aws_instance",
			Changes: []AttributeChange{
				{Attribute: "instance_type", StateValue: "t2.micro", CloudValue: "t3.medium"},
			},
		},
		{ResourceID: "i-999", ResourceType: "aws_instance", OnlyInState: true},
		{ResourceID: "i-888", ResourceType: "aws_instance", OnlyInCloud: true},
	}

	out := FormatDiff(diffs)
	if !strings.Contains(out, "[~] i-123") {
		t.Errorf("expected drift marker for i-123, got:\n%s", out)
	}
	if !strings.Contains(out, "[-] i-999") {
		t.Errorf("expected only-in-state marker for i-999, got:\n%s", out)
	}
	if !strings.Contains(out, "[+] i-888") {
		t.Errorf("expected only-in-cloud marker for i-888, got:\n%s", out)
	}
	if !strings.Contains(out, `state="t2.micro"`) {
		t.Errorf("expected state value in output, got:\n%s", out)
	}
}
