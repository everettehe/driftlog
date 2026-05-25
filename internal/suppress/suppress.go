// Package suppress provides functionality to suppress known drift entries
// so they are excluded from reports and alerts.
package suppress

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/driftlog/internal/diff"
)

// Rule defines a suppression rule that matches drift results to ignore.
type Rule struct {
	ResourceID   string `json:"resource_id"`   // exact match or "*" wildcard
	ResourceType string `json:"resource_type"` // exact match or "*" wildcard
	Attribute    string `json:"attribute"`     // attribute name or "*" for all
	Reason       string `json:"reason"`
	ExpiresAt    string `json:"expires_at"` // RFC3339 or empty for permanent
}

// List holds a collection of suppression rules.
type List struct {
	Rules []Rule `json:"rules"`
}

// LoadFile reads suppression rules from a JSON file.
// If the file does not exist, an empty List is returned.
func LoadFile(path string) (*List, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &List{}, nil
	}
	if err != nil {
		return nil, err
	}
	var l List
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

// Apply filters out drift results (or individual attribute changes) that match
// any active suppression rule. Results with no remaining changes are removed.
func Apply(results []diff.Result, l *List) []diff.Result {
	if l == nil || len(l.Rules) == 0 {
		return results
	}
	now := time.Now()
	var out []diff.Result
	for _, r := range results {
		filtered := filterResult(r, l.Rules, now)
		if filtered != nil {
			out = append(out, *filtered)
		}
	}
	return out
}

func filterResult(r diff.Result, rules []Rule, now time.Time) *diff.Result {
	var kept []diff.AttributeChange
	for _, ch := range r.Changes {
		if !isSuppressed(r, ch.Attribute, rules, now) {
			kept = append(kept, ch)
		}
	}
	// If the entire resource is suppressed (no changes left and it was drifted)
	if len(r.Changes) > 0 && len(kept) == 0 {
		return nil
	}
	// Suppress whole-resource statuses (OnlyInState / OnlyInCloud)
	if len(r.Changes) == 0 && isSuppressed(r, "*", rules, now) {
		return nil
	}
	copy := r
	copy.Changes = kept
	return &copy
}

func isSuppressed(r diff.Result, attribute string, rules []Rule, now time.Time) bool {
	for _, rule := range rules {
		if rule.ExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, rule.ExpiresAt)
			if err != nil || now.After(t) {
				continue
			}
		}
		if !matchField(rule.ResourceID, r.ResourceID) {
			continue
		}
		if !matchField(rule.ResourceType, r.ResourceType) {
			continue
		}
		if rule.Attribute == "*" || strings.EqualFold(rule.Attribute, attribute) {
			return true
		}
	}
	return false
}

func matchField(pattern, value string) bool {
	return pattern == "*" || strings.EqualFold(pattern, value)
}
