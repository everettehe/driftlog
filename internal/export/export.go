// Package export provides functionality to export drift results to various file formats.
package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/yourorg/driftlog/internal/diff"
)

// Format represents a supported export format.
type Format string

const (
	FormatCSV  Format = "csv"
	FormatJSON Format = "json"
	FormatNDJSON Format = "ndjson"
)

// ParseFormat parses a string into a Format, returning an error if unsupported.
func ParseFormat(s string) (Format, error) {
	switch Format(strings.ToLower(s)) {
	case FormatCSV, FormatJSON, FormatNDJSON:
		return Format(strings.ToLower(s)), nil
	}
	return "", fmt.Errorf("unsupported export format %q: must be csv, json, or ndjson", s)
}

// Write serialises results to w using the given format.
func Write(w io.Writer, results []diff.Result, format Format) error {
	switch format {
	case FormatCSV:
		return writeCSV(w, results)
	case FormatJSON:
		return writeJSON(w, results)
	case FormatNDJSON:
		return writeNDJSON(w, results)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func writeCSV(w io.Writer, results []diff.Result) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"resource_id", "resource_type", "status", "attribute", "state_value", "cloud_value"}); err != nil {
		return err
	}
	for _, r := range results {
		if len(r.Changes) == 0 {
			if err := cw.Write([]string{r.ResourceID, r.ResourceType, string(r.Status), "", "", ""}); err != nil {
				return err
			}
			continue
		}
		for _, c := range r.Changes {
			if err := cw.Write([]string{r.ResourceID, r.ResourceType, string(r.Status), c.Attribute, c.StateValue, c.CloudValue}); err != nil {
				return err
			}
		}
	}
	cw.Flush()
	return cw.Error()
}

func writeJSON(w io.Writer, results []diff.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func writeNDJSON(w io.Writer, results []diff.Result) error {
	enc := json.NewEncoder(w)
	for _, r := range results {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}
