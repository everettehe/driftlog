// Package watchlist manages a prioritised list of resources to monitor closely.
package watchlist

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/driftlog/internal/diff"
)

// Entry represents a single watched resource.
type Entry struct {
	ResourceID string    `json:"resource_id"`
	ResourceType string  `json:"resource_type"`
	Reason     string    `json:"reason"`
	AddedAt    time.Time `json:"added_at"`
}

// Watchlist holds a collection of watched resource entries.
type Watchlist struct {
	Entries []Entry `json:"entries"`
}

// LoadFile reads a watchlist from a JSON file. Returns an empty Watchlist if
// the file does not exist.
func LoadFile(path string) (*Watchlist, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Watchlist{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("watchlist: read file: %w", err)
	}
	var wl Watchlist
	if err := json.Unmarshal(data, &wl); err != nil {
		return nil, fmt.Errorf("watchlist: parse JSON: %w", err)
	}
	return &wl, nil
}

// SaveFile persists the watchlist to a JSON file.
func (wl *Watchlist) SaveFile(path string) error {
	data, err := json.MarshalIndent(wl, "", "  ")
	if err != nil {
		return fmt.Errorf("watchlist: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("watchlist: write file: %w", err)
	}
	return nil
}

// Add appends a new entry if the resource is not already watched.
func (wl *Watchlist) Add(resourceID, resourceType, reason string) bool {
	for _, e := range wl.Entries {
		if e.ResourceID == resourceID {
			return false
		}
	}
	wl.Entries = append(wl.Entries, Entry{
		ResourceID:   resourceID,
		ResourceType: resourceType,
		Reason:       reason,
		AddedAt:      time.Now().UTC(),
	})
	return true
}

// Remove deletes an entry by resource ID. Returns true if an entry was removed.
func (wl *Watchlist) Remove(resourceID string) bool {
	for i, e := range wl.Entries {
		if e.ResourceID == resourceID {
			wl.Entries = append(wl.Entries[:i], wl.Entries[i+1:]...)
			return true
		}
	}
	return false
}

// Filter returns only the diff results whose resource ID appears in the watchlist.
func (wl *Watchlist) Filter(results []diff.Result) []diff.Result {
	index := make(map[string]struct{}, len(wl.Entries))
	for _, e := range wl.Entries {
		index[strings.ToLower(e.ResourceID)] = struct{}{}
	}
	var out []diff.Result
	for _, r := range results {
		if _, ok := index[strings.ToLower(r.ResourceID)]; ok {
			out = append(out, r)
		}
	}
	return out
}
