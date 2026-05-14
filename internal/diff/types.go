package diff

// DriftStatus describes the relationship between a state resource and live cloud.
type DriftStatus string

const (
	// StatusMatch means the live resource matches the state exactly.
	StatusMatch DriftStatus = "match"
	// StatusDrifted means the resource exists in both but attributes differ.
	StatusDrifted DriftStatus = "drifted"
	// StatusMissing means the resource is in state but not found in the cloud.
	StatusMissing DriftStatus = "missing"
	// StatusOrphaned means the resource exists in the cloud but not in state.
	StatusOrphaned DriftStatus = "orphaned"
)

// DriftResult captures the comparison outcome for a single resource.
type DriftResult struct {
	ResourceID   string
	ResourceType string
	Status       DriftStatus
	// Diffs holds human-readable lines describing each attribute change.
	Diffs []string
}
