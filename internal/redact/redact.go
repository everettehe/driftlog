// Package redact provides utilities for scrubbing sensitive attribute values
// from drift results before they are written to output or notifications.
package redact

import (
	"strings"

	"github.com/example/driftlog/internal/diff"
)

// DefaultSensitiveKeys contains attribute key substrings that are considered
// sensitive and should be redacted from output.
var DefaultSensitiveKeys = []string{
	"password",
	"secret",
	"token",
	"private_key",
	"access_key",
	"api_key",
	"credentials",
}

const redactedValue = "[REDACTED]"

// Apply returns a copy of the provided results with sensitive attribute values
// replaced by a placeholder string. Keys are matched case-insensitively against
// the provided sensitiveKeys list (or DefaultSensitiveKeys when nil).
func Apply(results []diff.Result, sensitiveKeys []string) []diff.Result {
	if sensitiveKeys == nil {
		sensitiveKeys = DefaultSensitiveKeys
	}

	out := make([]diff.Result, len(results))
	for i, r := range results {
		out[i] = redactResult(r, sensitiveKeys)
	}
	return out
}

func redactResult(r diff.Result, sensitiveKeys []string) diff.Result {
	redacted := make([]diff.AttributeDiff, len(r.Diffs))
	for i, d := range r.Diffs {
		if isSensitive(d.Key, sensitiveKeys) {
			d.StateValue = redactedValue
			d.CloudValue = redactedValue
		}
		redacted[i] = d
	}
	r.Diffs = redacted
	return r
}

func isSensitive(key string, sensitiveKeys []string) bool {
	lower := strings.ToLower(key)
	for _, s := range sensitiveKeys {
		if strings.Contains(lower, strings.ToLower(s)) {
			return true
		}
	}
	return false
}
