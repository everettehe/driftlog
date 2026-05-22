// Package tags provides utilities for comparing resource tags between
// Terraform state and live cloud resources, surfacing tag-level drift.
package tags

import (
	"fmt"
	"sort"
	"strings"
)

// Diff represents a single tag-level difference.
type Diff struct {
	Key      string
	Expected string // value from Terraform state; empty if missing
	Actual   string // value from live cloud; empty if missing
	Status   string // "added", "removed", or "changed"
}

// Compare compares two tag maps (state vs live) and returns a slice of Diff
// entries for any keys that differ. Tags present in both maps with equal
// values are omitted.
func Compare(state, live map[string]string) []Diff {
	seen := make(map[string]bool)
	var diffs []Diff

	for k, sv := range state {
		seen[k] = true
		lv, ok := live[k]
		if !ok {
			diffs = append(diffs, Diff{Key: k, Expected: sv, Actual: "", Status: "removed"})
		} else if sv != lv {
			diffs = append(diffs, Diff{Key: k, Expected: sv, Actual: lv, Status: "changed"})
		}
	}

	for k, lv := range live {
		if !seen[k] {
			diffs = append(diffs, Diff{Key: k, Expected: "", Actual: lv, Status: "added"})
		}
	}

	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Key < diffs[j].Key
	})
	return diffs
}

// Format returns a human-readable summary of tag diffs, one line per diff.
// Returns an empty string when there are no diffs.
func Format(diffs []Diff) string {
	if len(diffs) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, d := range diffs {
		switch d.Status {
		case "added":
			fmt.Fprintf(&sb, "  + tag %q = %q (not in state)\n", d.Key, d.Actual)
		case "removed":
			fmt.Fprintf(&sb, "  - tag %q = %q (missing in cloud)\n", d.Key, d.Expected)
		case "changed":
			fmt.Fprintf(&sb, "  ~ tag %q: %q → %q\n", d.Key, d.Expected, d.Actual)
		}
	}
	return sb.String()
}

// HasDrift returns true when the two tag maps differ in any way.
func HasDrift(state, live map[string]string) bool {
	return len(Compare(state, live)) > 0
}
