package policy

import (
	"fmt"
	"strings"

	"github.com/driftlog/internal/diff"
)

// NoDriftAllowed returns a Rule that flags any resource with drift as an error.
func NoDriftAllowed() Rule {
	return Rule{
		Name:        "no-drift-allowed",
		Description: "All resources must match Terraform state exactly.",
		Severity:    "error",
		Check: func(r diff.Result) *Violation {
			if r.Status == diff.StatusDrifted {
				return &Violation{
					Message: fmt.Sprintf("%d attribute(s) differ from state", len(r.Changes)),
				}
			}
			return nil
		},
	}
}

// NoUnmanagedResources returns a Rule that flags resources present in cloud
// but absent from Terraform state.
func NoUnmanagedResources() Rule {
	return Rule{
		Name:        "no-unmanaged-resources",
		Description: "Cloud resources must be managed by Terraform.",
		Severity:    "warn",
		Check: func(r diff.Result) *Violation {
			if r.Status == diff.StatusOnlyInCloud {
				return &Violation{
					Message: "resource exists in cloud but not in Terraform state",
				}
			}
			return nil
		},
	}
}

// RequireAttribute returns a Rule that ensures a specific attribute is present
// and non-empty on drifted or cloud-only resources.
func RequireAttribute(attr string) Rule {
	return Rule{
		Name:        fmt.Sprintf("require-attribute/%s", attr),
		Description: fmt.Sprintf("Resources must have attribute '%s'.", attr),
		Severity:    "warn",
		Check: func(r diff.Result) *Violation {
			if r.Status == diff.StatusMatch {
				return nil
			}
			val, ok := r.CloudAttributes[attr]
			if !ok || strings.TrimSpace(fmt.Sprintf("%v", val)) == "" {
				return &Violation{
					Message: fmt.Sprintf("missing or empty attribute '%s'", attr),
				}
			}
			return nil
		},
	}
}

// DefaultRules returns the standard set of built-in policy rules.
func DefaultRules() []Rule {
	return []Rule{
		NoDriftAllowed(),
		NoUnmanagedResources(),
	}
}
