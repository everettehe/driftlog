package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/user/driftlog/internal/report"
	"github.com/user/driftlog/internal/diff"
)

// Format represents the supported output formats.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// ParseFormat converts a string to a Format, returning an error if unsupported.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "text", "":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported output format %q: must be \"text\" or \"json\"", s)
	}
}

// Write writes drift results to w using the specified format.
// If w is nil, os.Stdout is used.
func Write(w io.Writer, results []diff.DriftResult, format Format) error {
	if w == nil {
		w = os.Stdout
	}
	switch format {
	case FormatJSON:
		return writeJSON(w, results)
	default:
		return report.WriteText(w, results)
	}
}

// writeJSON encodes results as a JSON array to w.
func writeJSON(w io.Writer, results []diff.DriftResult) error {
	enc := newJSONEncoder(w)
	return enc.Encode(results)
}

// newJSONEncoder returns a configured JSON encoder.
func newJSONEncoder(w io.Writer) *jsonEncoder {
	return &jsonEncoder{w: w}
}

type jsonEncoder struct {
	w io.Writer
}

func (e *jsonEncoder) Encode(results []diff.DriftResult) error {
	data, err := marshalResults(results)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	_, err = fmt.Fprintln(e.w, string(data))
	return err
}
