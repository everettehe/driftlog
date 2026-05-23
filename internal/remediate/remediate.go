package remediate

import (
	"fmt"
	"strings"

	"github.com/driftlog/internal/diff"
)

// Action represents a suggested remediation action for a drifted resource.
type Action struct {
	ResourceID string
	ResourceType string
	Command string
	Description string
}

// Suggest generates remediation actions for drifted resources.
func Suggest(results []diff.Result) []Action {
	var actions []Action
	for _, r := range results {
		switch r.Status {
		case diff.StatusDrifted:
			actions = append(actions, Action{
				ResourceID: r.ResourceID,
				ResourceType: r.ResourceType,
				Command: buildTerraformApply(r.ResourceID),
				Description: fmt.Sprintf("Resource has %d attribute(s) out of sync with Terraform state.", len(r.Changes)),
			})
		case diff.StatusOnlyInCloud:
			actions = append(actions, Action{
				ResourceID: r.ResourceID,
				ResourceType: r.ResourceType,
				Command: buildTerraformImport(r.ResourceType, r.ResourceID),
				Description: "Resource exists in cloud but is not tracked in Terraform state. Consider importing.",
			})
		case diff.StatusOnlyInState:
			actions = append(actions, Action{
				ResourceID: r.ResourceID,
				ResourceType: r.ResourceType,
				Command: buildTerraformDestroy(r.ResourceID),
				Description: "Resource is in Terraform state but missing from cloud. It may have been deleted manually.",
			})
		}
	}
	return actions
}

// Format returns a human-readable string of all remediation actions.
func Format(actions []Action) string {
	if len(actions) == 0 {
		return "No remediation actions required.\n"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Remediation suggestions (%d):\n\n", len(actions)))
	for _, a := range actions {
		sb.WriteString(fmt.Sprintf("  [%s] %s\n", a.ResourceType, a.ResourceID))
		sb.WriteString(fmt.Sprintf("  → %s\n", a.Description))
		sb.WriteString(fmt.Sprintf("  $ %s\n\n", a.Command))
	}
	return sb.String()
}

func buildTerraformApply(resourceID string) string {
	return fmt.Sprintf("terraform apply -target=%s", resourceID)
}

func buildTerraformImport(resourceType, resourceID string) string {
	return fmt.Sprintf("terraform import %s %s", resourceType, resourceID)
}

func buildTerraformDestroy(resourceID string) string {
	return fmt.Sprintf("terraform destroy -target=%s", resourceID)
}
