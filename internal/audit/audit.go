// Package audit records drift scan events to an append-only audit log.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/your-org/driftlog/internal/diff"
)

// Entry represents a single audit log record.
type Entry struct {
	Timestamp  time.Time `json:"timestamp"`
	RunID      string    `json:"run_id"`
	TotalCount int       `json:"total_count"`
	DriftCount int       `json:"drift_count"`
	CleanCount int       `json:"clean_count"`
	MissingCount int    `json:"missing_count"`
	UnmanagedCount int  `json:"unmanaged_count"`
	StateFile  string    `json:"state_file"`
	Region     string    `json:"region"`
}

// Append writes a new audit entry to the given file path (JSONL format).
func Append(path string, results []diff.Result, runID, stateFile, region string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("audit: open %s: %w", path, err)
	}
	defer f.Close()

	entry := build(results, runID, stateFile, region)
	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}
	_, err = fmt.Fprintf(f, "%s\n", line)
	return err
}

// ReadAll reads all audit entries from the given JSONL file.
func ReadAll(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("audit: read %s: %w", path, err)
	}

	var entries []Entry
	dec := json.NewDecoder(
		// wrap bytes in a reader via strings trick
		newBytesReader(data),
	)
	for dec.More() {
		var e Entry
		if err := dec.Decode(&e); err != nil {
			return nil, fmt.Errorf("audit: decode entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func build(results []diff.Result, runID, stateFile, region string) Entry {
	e := Entry{
		Timestamp: time.Now().UTC(),
		RunID:     runID,
		StateFile: stateFile,
		Region:    region,
		TotalCount: len(results),
	}
	for _, r := range results {
		switch r.Status {
		case diff.StatusDrifted:
			e.DriftCount++
		case diff.StatusClean:
			e.CleanCount++
		case diff.StatusOnlyInState:
			e.MissingCount++
		case diff.StatusOnlyInCloud:
			e.UnmanagedCount++
		}
	}
	return e
}
