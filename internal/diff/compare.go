package diff

import (
	"fmt"
	"sort"
	"strings"
)

// ResourceDiff represents the drift between a Terraform state resource and a live cloud resource.
type ResourceDiff struct {
	ResourceID   string
	ResourceType string
	Changes      []AttributeChange
	OnlyInState  bool
	OnlyInCloud  bool
}

// AttributeChange represents a single attribute that has drifted.
type AttributeChange struct {
	Attribute string
	StateValue string
	CloudValue string
}

// Resource is a minimal interface for state and cloud resources.
type Resource struct {
	ID         string
	Type       string
	Attributes map[string]string
}

// Compare takes a slice of state resources and a slice of live cloud resources
// and returns a list of ResourceDiffs describing any drift.
func Compare(stateResources, cloudResources []Resource) []ResourceDiff {
	var diffs []ResourceDiff

	cloudMap := make(map[string]Resource, len(cloudResources))
	for _, r := range cloudResources {
		cloudMap[r.ID] = r
	}

	seenIDs := make(map[string]bool)

	for _, sr := range stateResources {
		seenIDs[sr.ID] = true
		cr, exists := cloudMap[sr.ID]
		if !exists {
			diffs = append(diffs, ResourceDiff{
				ResourceID:   sr.ID,
				ResourceType: sr.Type,
				OnlyInState:  true,
			})
			continue
		}

		changes := compareAttributes(sr.Attributes, cr.Attributes)
		if len(changes) > 0 {
			diffs = append(diffs, ResourceDiff{
				ResourceID:   sr.ID,
				ResourceType: sr.Type,
				Changes:      changes,
			})
		}
	}

	for _, cr := range cloudResources {
		if !seenIDs[cr.ID] {
			diffs = append(diffs, ResourceDiff{
				ResourceID:   cr.ID,
				ResourceType: cr.Type,
				OnlyInCloud:  true,
			})
		}
	}

	return diffs
}

func compareAttributes(state, cloud map[string]string) []AttributeChange {
	var changes []AttributeChange
	keys := make(map[string]struct{})
	for k := range state {
		keys[k] = struct{}{}
	}
	for k := range cloud {
		keys[k] = struct{}{}
	}

	sortedKeys := make([]string, 0, len(keys))
	for k := range keys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		sv := state[k]
		cv := cloud[k]
		if sv != cv {
			changes = append(changes, AttributeChange{
				Attribute:  k,
				StateValue: sv,
				CloudValue: cv,
			})
		}
	}
	return changes
}

// FormatDiff returns a human-readable string for a slice of ResourceDiffs.
func FormatDiff(diffs []ResourceDiff) string {
	if len(diffs) == 0 {
		return "No drift detected."
	}
	var sb strings.Builder
	for _, d := range diffs {
		switch {
		case d.OnlyInState:
			sb.WriteString(fmt.Sprintf("[-] %s (%s): exists in state but not found in cloud\n", d.ResourceID, d.ResourceType))
		case d.OnlyInCloud:
			sb.WriteString(fmt.Sprintf("[+] %s (%s): exists in cloud but not in state\n", d.ResourceID, d.ResourceType))
		default:
			sb.WriteString(fmt.Sprintf("[~] %s (%s): attribute drift\n", d.ResourceID, d.ResourceType))
			for _, c := range d.Changes {
				sb.WriteString(fmt.Sprintf("      %s: state=%q cloud=%q\n", c.Attribute, c.StateValue, c.CloudValue))
			}
		}
	}
	return sb.String()
}
