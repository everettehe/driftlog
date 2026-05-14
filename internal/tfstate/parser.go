package tfstate

import (
	"encoding/json"
	"fmt"
	"os"
)

// Resource represents a single resource extracted from Terraform state.
type Resource struct {
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	Provider   string            `json:"provider"`
	Attributes map[string]any    `json:"attributes"`
}

// State holds the parsed contents of a Terraform state file.
type State struct {
	Version   int        `json:"version"`
	Resources []Resource `json:"resources"`
}

type tfStateFile struct {
	Version   int              `json:"version"`
	Resources []tfStateResource `json:"resources"`
}

type tfStateResource struct {
	Type      string            `json:"type"`
	Name      string            `json:"name"`
	Provider  string            `json:"provider"`
	Instances []tfStateInstance `json:"instances"`
}

type tfStateInstance struct {
	Attributes map[string]any `json:"attributes"`
}

// ParseFile reads and parses a Terraform state file from the given path.
func ParseFile(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}
	return Parse(data)
}

// Parse parses raw Terraform state JSON bytes into a State struct.
func Parse(data []byte) (*State, error) {
	var raw tfStateFile
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshalling state: %w", err)
	}

	if raw.Version < 4 {
		return nil, fmt.Errorf("unsupported state version %d (minimum: 4)", raw.Version)
	}

	state := &State{Version: raw.Version}
	for _, r := range raw.Resources {
		attrs := map[string]any{}
		if len(r.Instances) > 0 {
			attrs = r.Instances[0].Attributes
		}
		state.Resources = append(state.Resources, Resource{
			Type:       r.Type,
			Name:       r.Name,
			Provider:   r.Provider,
			Attributes: attrs,
		})
	}
	return state, nil
}
