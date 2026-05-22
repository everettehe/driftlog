// Package history records drift scan results over time and provides
// trend analysis across multiple runs.
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/example/driftlog/internal/diff"
)

// Entry represents a single recorded scan run.
type Entry struct {
	Timestamp  time.Time         `json:"timestamp"`
	TotalCount int               `json:"total_count"`
	DriftCount int               `json:"drift_count"`
	Results    []diff.DriftResult `json:"results"`
}

// History holds a collection of scan entries.
type History struct {
	Entries []Entry `json:"entries"`
}

// Store appends a new entry derived from the given results to the history file at path.
func Store(path string, results []diff.DriftResult) error {
	h, err := load(path)
	if err != nil {
		return fmt.Errorf("history: load: %w", err)
	}

	drifted := 0
	for _, r := range results {
		if r.Status != diff.StatusMatch {
			drifted++
		}
	}

	h.Entries = append(h.Entries, Entry{
		Timestamp:  time.Now().UTC(),
		TotalCount: len(results),
		DriftCount: drifted,
		Results:    results,
	})

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("history: marshal: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// Load reads and returns the history from the file at path.
// Returns an empty History if the file does not exist.
func Load(path string) (*History, error) {
	return load(path)
}

// Trend returns the last n entries sorted by timestamp ascending.
func (h *History) Trend(n int) []Entry {
	sorted := make([]Entry, len(h.Entries))
	copy(sorted, h.Entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})
	if n > 0 && len(sorted) > n {
		return sorted[len(sorted)-n:]
	}
	return sorted
}

func load(path string) (*History, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &History{}, nil
	}
	if err != nil {
		return nil, err
	}
	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("invalid history JSON: %w", err)
	}
	return &h, nil
}
