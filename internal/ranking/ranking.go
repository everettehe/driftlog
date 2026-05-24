package ranking

import (
	"sort"

	"github.com/yourorg/driftlog/internal/diff"
)

// Entry holds a resource with its computed drift score.
type Entry struct {
	ResourceID string
	ResourceType string
	Status       diff.Status
	ChangeCount  int
	Score        int
}

// Rank returns resources sorted by drift severity (highest score first).
// Scoring: OnlyInCloud=10, OnlyInState=8, Drifted=changeCount*3, Clean=0.
func Rank(results []diff.Result) []Entry {
	entries := make([]Entry, 0, len(results))
	for _, r := range results {
		entries = append(entries, Entry{
			ResourceID:   r.ResourceID,
			ResourceType: r.ResourceType,
			Status:       r.Status,
			ChangeCount:  len(r.Changes),
			Score:        computeScore(r),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Score != entries[j].Score {
			return entries[i].Score > entries[j].Score
		}
		return entries[i].ResourceID < entries[j].ResourceID
	})
	return entries
}

// TopN returns at most n highest-ranked entries.
func TopN(results []diff.Result, n int) []Entry {
	all := Rank(results)
	if n <= 0 || n >= len(all) {
		return all
	}
	return all[:n]
}

func computeScore(r diff.Result) int {
	switch r.Status {
	case diff.StatusOnlyInCloud:
		return 10
	case diff.StatusOnlyInState:
		return 8
	case diff.StatusDrifted:
		return len(r.Changes) * 3
	default:
		return 0
	}
}
