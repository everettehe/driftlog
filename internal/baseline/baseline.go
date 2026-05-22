// Package baseline provides functionality to capture and compare drift snapshots
// against a known-good baseline state.
package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/yourorg/driftlog/internal/diff"
)

// Snapshot represents a captured baseline of drift results at a point in time.
type Snapshot struct {
	CapturedAt time.Time          `json:"captured_at"`
	Results    []diff.DriftResult `json:"results"`
}

// Save writes a baseline snapshot to the given file path.
func Save(path string, results []diff.DriftResult) error {
	snap := Snapshot{
		CapturedAt: time.Now().UTC(),
		Results:    results,
	}

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("baseline: marshal snapshot: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("baseline: write file %q: %w", path, err)
	}

	return nil
}

// Load reads a baseline snapshot from the given file path.
// Returns nil, nil if the file does not exist.
func Load(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return fmt.Errorf("baseline: read file %q: %w", path, err), nil
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("baseline: unmarshal snapshot: %w", err)
	}

	return &snap, nil
}

// NewDrifts returns only the DriftResults that are new compared to the baseline.
// A result is considered new if its ResourceID did not appear as drifted in the baseline.
func NewDrifts(baseline *Snapshot, current []diff.DriftResult) []diff.DriftResult {
	if baseline == nil {
		return current
	}

	known := make(map[string]bool, len(baseline.Results))
	for _, r := range baseline.Results {
		if r.Status == diff.StatusDrifted {
			known[r.ResourceID] = true
		}
	}

	var novel []diff.DriftResult
	for _, r := range current {
		if r.Status == diff.StatusDrifted && !known[r.ResourceID] {
			novel = append(novel, r)
		}
	}

	return novel
}
