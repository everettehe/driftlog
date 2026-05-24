package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/yourorg/driftlog/internal/diff"
	"github.com/yourorg/driftlog/internal/groupby"
)

// runGroupBy groups drift results by the given dimension and prints a summary.
// dimension is one of: type, status, tag:<key>
func runGroupBy(results []diff.Result, dimension string) error {
	var groups []groupby.Group

	switch {
	case dimension == "type":
		groups = groupby.ByType(results)
	case dimension == "status":
		groups = groupby.ByStatus(results)
	case strings.HasPrefix(dimension, "tag:"):
		tagKey := strings.TrimPrefix(dimension, "tag:")
		if tagKey == "" {
			return fmt.Errorf("tag key must not be empty (use tag:<key>)")
		}
		groups = groupby.ByTag(results, tagKey)
	default:
		return fmt.Errorf("unknown group-by dimension %q; use type, status, or tag:<key>", dimension)
	}

	if len(groups) == 0 {
		fmt.Fprintln(os.Stdout, "No results to group.")
		return nil
	}

	for _, g := range groups {
		drifted := 0
		for _, r := range g.Results {
			if r.Status == diff.StatusDrifted {
				drifted++
			}
		}
		fmt.Fprintf(os.Stdout, "%-40s  total=%-4d  drifted=%d\n",
			g.Label, len(g.Results), drifted)
	}
	return nil
}
