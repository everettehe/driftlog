// Package groupby provides utilities for grouping drift results by various dimensions.
package groupby

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yourorg/driftlog/internal/diff"
)

// Group holds a label and the results belonging to that group.
type Group struct {
	Label   string
	Results []diff.Result
}

// ByType groups results by their resource type.
func ByType(results []diff.Result) []Group {
	return groupBy(results, func(r diff.Result) string {
		return r.ResourceType
	})
}

// ByStatus groups results by their drift status.
func ByStatus(results []diff.Result) []Group {
	return groupBy(results, func(r diff.Result) string {
		return string(r.Status)
	})
}

// ByTag groups results by the value of a given tag key.
// Resources missing the tag are placed under the label "<untagged>".
func ByTag(results []diff.Result, tagKey string) []Group {
	norm := strings.ToLower(tagKey)
	return groupBy(results, func(r diff.Result) string {
		for k, v := range r.Attributes {
			if strings.ToLower(k) == norm {
				return fmt.Sprintf("%s=%s", tagKey, v)
			}
		}
		return "<untagged>"
	})
}

// groupBy is the generic implementation that partitions results using keyFn
// and returns groups sorted alphabetically by label.
func groupBy(results []diff.Result, keyFn func(diff.Result) string) []Group {
	index := make(map[string][]diff.Result)
	order := []string{}

	for _, r := range results {
		key := keyFn(r)
		if _, exists := index[key]; !exists {
			order = append(order, key)
		}
		index[key] = append(index[key], r)
	}

	sort.Strings(order)

	groups := make([]Group, 0, len(order))
	for _, label := range order {
		groups = append(groups, Group{
			Label:   label,
			Results: index[label],
		})
	}
	return groups
}
