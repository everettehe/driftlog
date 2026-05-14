package aws

import (
	"testing"
)

func TestBuildResourcesFromInstances_Basic(t *testing.T) {
	instances := []mockInstance{
		{ID: "i-abc123", InstanceType: "t3.micro", State: "running", AMI: "ami-0abcdef"},
		{ID: "i-def456", InstanceType: "t3.small", State: "stopped", AMI: "ami-1234567"},
	}

	resources := buildResourcesFromInstances(instances)

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	if resources[0].Type != "aws_instance" {
		t.Errorf("expected type aws_instance, got %s", resources[0].Type)
	}
	if resources[0].ID != "i-abc123" {
		t.Errorf("expected ID i-abc123, got %s", resources[0].ID)
	}
	if resources[0].Attributes["instance_type"] != "t3.micro" {
		t.Errorf("expected instance_type t3.micro, got %s", resources[0].Attributes["instance_type"])
	}
	if resources[1].Attributes["state"] != "stopped" {
		t.Errorf("expected state stopped, got %s", resources[1].Attributes["state"])
	}
}

func TestBuildResourcesFromInstances_Empty(t *testing.T) {
	resources := buildResourcesFromInstances(nil)
	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

func TestBuildResourcesFromBuckets_Basic(t *testing.T) {
	buckets := []mockBucket{
		{Name: "my-logs-bucket"},
		{Name: "my-assets-bucket"},
	}

	resources := buildResourcesFromBuckets(buckets)

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}
	if resources[0].Type != "aws_s3_bucket" {
		t.Errorf("expected type aws_s3_bucket, got %s", resources[0].Type)
	}
	if resources[0].ID != "my-logs-bucket" {
		t.Errorf("expected ID my-logs-bucket, got %s", resources[0].ID)
	}
	if resources[0].Attributes["bucket"] != "my-logs-bucket" {
		t.Errorf("expected bucket attribute my-logs-bucket, got %s", resources[0].Attributes["bucket"])
	}
}

func TestBuildResourcesFromBuckets_Empty(t *testing.T) {
	resources := buildResourcesFromBuckets(nil)
	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

func TestResource_AttributeKeys(t *testing.T) {
	r := Resource{
		Type: "aws_instance",
		ID:   "i-test",
		Attributes: map[string]string{
			"instance_type": "t2.micro",
			"state":         "running",
		},
	}
	if _, ok := r.Attributes["instance_type"]; !ok {
		t.Error("expected instance_type attribute to be present")
	}
}
