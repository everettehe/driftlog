// Package snapshot captures and compares full drift scan results over time,
// allowing users to detect when drift was introduced or resolved.
package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yourorg/driftlog/internal/diff"
)

// Snapshot represents a point-in-time capture of drift scan results.
type Snapshot struct {
	ID        string            `json:"id"`
	CreatedAt time.Time         `json:"created_at"`
	Label     string            `json:"label,omitempty"`
	Results   []diff.DriftResult `json:"results"`
}

// Save writes a snapshot to the given directory with a timestamped filename.
func Save(dir string, results []diff.DriftResult, label string) (*Snapshot, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("snapshot: create dir: %w", err)
	}

	snap := &Snapshot{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		CreatedAt: time.Now().UTC(),
		Label:     label,
		Results:   results,
	}

	filename := fmt.Sprintf("%s.json", snap.ID)
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("snapshot: marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return nil, fmt.Errorf("snapshot: write file: %w", err)
	}

	return snap, nil
}

// Load reads a snapshot from the given file path.
func Load(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("snapshot: read file: %w", err)
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("snapshot: unmarshal: %w", err)
	}

	return &snap, nil
}

// Diff returns resources whose drift status changed between two snapshots.
func Diff(previous, current *Snapshot) []StatusChange {
	prevIndex := indexByID(previous.Results)
	var changes []StatusChange

	for _, cur := range current.Results {
		prev, existed := prevIndex[cur.ResourceID]
		if !existed || prev.Status != cur.Status {
			changes = append(changes, StatusChange{
				ResourceID:   cur.ResourceID,
				ResourceType: cur.ResourceType,
				PrevStatus:   statusOrNew(existed, prev),
				CurrStatus:   cur.Status,
			})
		}
	}

	return changes
}

// StatusChange describes a resource whose drift status changed between snapshots.
type StatusChange struct {
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	PrevStatus   string `json:"prev_status"`
	CurrStatus   string `json:"curr_status"`
}

func indexByID(results []diff.DriftResult) map[string]diff.DriftResult {
	m := make(map[string]diff.DriftResult, len(results))
	for _, r := range results {
		m[r.ResourceID] = r
	}
	return m
}

func statusOrNew(existed bool, r diff.DriftResult) string {
	if !existed {
		return "new"
	}
	return r.Status
}
