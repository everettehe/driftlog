package tfstate_test

import (
	"testing"

	"github.com/yourorg/driftlog/internal/tfstate"
)

var validStateJSON = []byte(`{
  "version": 4,
  "terraform_version": "1.5.0",
  "resources": [
    {
      "type": "aws_instance",
      "name": "web",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "attributes": {
            "id": "i-0abc123",
            "instance_type": "t3.micro",
            "ami": "ami-0deadbeef"
          }
        }
      ]
    }
  ]
}`)

func TestParse_ValidState(t *testing.T) {
	state, err := tfstate.Parse(validStateJSON)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if state.Version != 4 {
		t.Errorf("expected version 4, got %d", state.Version)
	}
	if len(state.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(state.Resources))
	}
	res := state.Resources[0]
	if res.Type != "aws_instance" {
		t.Errorf("expected type aws_instance, got %s", res.Type)
	}
	if res.Name != "web" {
		t.Errorf("expected name web, got %s", res.Name)
	}
	if res.Attributes["instance_type"] != "t3.micro" {
		t.Errorf("unexpected instance_type: %v", res.Attributes["instance_type"])
	}
}

func TestParse_UnsupportedVersion(t *testing.T) {
	data := []byte(`{"version": 3, "resources": []}`)
	_, err := tfstate.Parse(data)
	if err == nil {
		t.Fatal("expected error for unsupported version, got nil")
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := tfstate.Parse([]byte(`not-json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParse_EmptyResources(t *testing.T) {
	data := []byte(`{"version": 4, "resources": []}`)
	state, err := tfstate.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(state.Resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(state.Resources))
	}
}
