package filter

import (
	"strings"

	"github.com/your-org/driftlog/internal/diff"
)

// Options holds filtering criteria for drift results.
type Options struct {
	// ResourceTypes limits results to specific resource types (e.g. "aws_instance").
	ResourceTypes []string
	// OnlyDrifted, when true, excludes resources with no drift.
	OnlyDrifted bool
	// ExcludeIDs is a set of resource IDs to suppress from output.
	ExcludeIDs []string
}

// Apply returns a filtered slice of DriftResult based on the given Options.
func Apply(results []diff.DriftResult, opts Options) []diff.DriftResult {
	excluded := make(map[string]struct{}, len(opts.ExcludeIDs))
	for _, id := range opts.ExcludeIDs {
		excluded[id] = struct{}{}
	}

	var filtered []diff.DriftResult
	for _, r := range results {
		if _, skip := excluded[r.ResourceID]; skip {
			continue
		}
		if opts.OnlyDrifted && r.Status == diff.StatusMatch {
			continue
		}
		if len(opts.ResourceTypes) > 0 && !matchesType(r.ResourceType, opts.ResourceTypes) {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered
}

// matchesType returns true if resourceType matches any entry in the list.
// Comparison is case-insensitive.
func matchesType(resourceType string, types []string) bool {
	rt := strings.ToLower(resourceType)
	for _, t := range types {
		if strings.ToLower(t) == rt {
			return true
		}
	}
	return false
}
