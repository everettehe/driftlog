package suppress_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/driftlog/internal/diff"
	"github.com/driftlog/internal/suppress"
)

func makeResult(id, rtype, status string, attrs ...string) diff.Result {
	var changes []diff.AttributeChange
	for i := 0; i+1 < len(attrs); i += 2 {
		changes = append(changes, diff.AttributeChange{
			Attribute: attrs[i],
			StateValue: "old",
			CloudValue: attrs[i+1],
		})
	}
	return diff.Result{ResourceID: id, ResourceType: rtype, Status: status, Changes: changes}
}

func writeTempList(t *testing.T, l suppress.List) string {
	t.Helper()
	data, _ := json.Marshal(l)
	path := filepath.Join(t.TempDir(), "suppress.json")
	_ = os.WriteFile(path, data, 0o644)
	return path
}

func TestApply_NoRules_ReturnsAll(t *testing.T) {
	results := []diff.Result{makeResult("i-1", "aws_instance", "drifted", "ami", "new-ami")}
	out := suppress.Apply(results, &suppress.List{})
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
}

func TestApply_SuppressExactResource(t *testing.T) {
	l := suppress.List{Rules: []suppress.Rule{
		{ResourceID: "i-1", ResourceType: "*", Attribute: "*", Reason: "known"},
	}}
	results := []diff.Result{makeResult("i-1", "aws_instance", "drifted", "ami", "new")}
	out := suppress.Apply(results, &l)
	if len(out) != 0 {
		t.Fatalf("expected 0 results, got %d", len(out))
	}
}

func TestApply_SuppressSpecificAttribute(t *testing.T) {
	l := suppress.List{Rules: []suppress.Rule{
		{ResourceID: "*", ResourceType: "aws_instance", Attribute: "ami", Reason: "ami rotation"},
	}}
	// two changes: ami (suppressed) and instance_type (kept)
	results := []diff.Result{makeResult("i-1", "aws_instance", "drifted", "ami", "new", "instance_type", "t3.small")}
	out := suppress.Apply(results, &l)
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if len(out[0].Changes) != 1 || out[0].Changes[0].Attribute != "instance_type" {
		t.Errorf("expected only instance_type change, got %+v", out[0].Changes)
	}
}

func TestApply_ExpiredRule_NotSuppressed(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	l := suppress.List{Rules: []suppress.Rule{
		{ResourceID: "*", ResourceType: "*", Attribute: "*", ExpiresAt: past},
	}}
	results := []diff.Result{makeResult("i-1", "aws_instance", "drifted", "ami", "new")}
	out := suppress.Apply(results, &l)
	if len(out) != 1 {
		t.Fatalf("expected result to remain after expired rule, got %d", len(out))
	}
}

func TestApply_FutureExpiry_Suppresses(t *testing.T) {
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	l := suppress.List{Rules: []suppress.Rule{
		{ResourceID: "*", ResourceType: "*", Attribute: "*", ExpiresAt: future},
	}}
	results := []diff.Result{makeResult("i-1", "aws_instance", "drifted", "ami", "new")}
	out := suppress.Apply(results, &l)
	if len(out) != 0 {
		t.Fatalf("expected suppression with future expiry, got %d", len(out))
	}
}

func TestLoadFile_MissingFile_ReturnsEmpty(t *testing.T) {
	l, err := suppress.LoadFile("/nonexistent/suppress.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(l.Rules) != 0 {
		t.Errorf("expected empty rules, got %d", len(l.Rules))
	}
}

func TestLoadFile_ValidJSON(t *testing.T) {
	list := suppress.List{Rules: []suppress.Rule{
		{ResourceID: "bucket-1", ResourceType: "aws_s3_bucket", Attribute: "acl", Reason: "legacy"},
	}}
	path := writeTempList(t, list)
	loaded, err := suppress.LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded.Rules) != 1 || loaded.Rules[0].ResourceID != "bucket-1" {
		t.Errorf("unexpected rules: %+v", loaded.Rules)
	}
}
