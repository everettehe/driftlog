// Package score computes a drift health score (0–100) for a set of diff results.
// 100 means fully in sync; lower scores indicate more drift.
package score

import (
	"fmt"
	"math"

	"github.com/example/driftlog/internal/diff"
)

// Result holds the computed score and a human-readable grade.
type Result struct {
	Score     int     // 0–100
	Grade     string  // A, B, C, D, F
	Total     int
	Drifted   int
	Missing   int
	Untracked int
}

// Compute calculates a health score from the provided diff results.
// Weights: drifted = 1 pt, missing (only-in-state) = 2 pts, untracked (only-in-cloud) = 1 pt.
func Compute(results []diff.Result) Result {
	total := len(results)
	if total == 0 {
		return Result{Score: 100, Grade: "A", Total: 0}
	}

	var drifted, missing, untracked int
	for _, r := range results {
		switch r.Status {
		case diff.StatusDrifted:
			drifted++
		case diff.StatusOnlyInState:
			missing++
		case diff.StatusOnlyInCloud:
			untracked++
		}
	}

	// Max possible penalty if every resource were missing (weight 2).
	maxPenalty := float64(total) * 2.0
	penalty := float64(drifted)*1.0 + float64(missing)*2.0 + float64(untracked)*1.0

	raw := 100.0 * (1.0 - penalty/maxPenalty)
	score := int(math.Round(math.Max(0, math.Min(100, raw))))

	return Result{
		Score:     score,
		Grade:     grade(score),
		Total:     total,
		Drifted:   drifted,
		Missing:   missing,
		Untracked: untracked,
	}
}

// Lines returns a human-readable summary suitable for CLI output.
func (r Result) Lines() []string {
	return []string{
		fmt.Sprintf("Health Score : %d / 100  (%s)", r.Score, r.Grade),
		fmt.Sprintf("Resources    : %d total, %d drifted, %d missing, %d untracked",
			r.Total, r.Drifted, r.Missing, r.Untracked),
	}
}

func grade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 75:
		return "B"
	case score >= 60:
		return "C"
	case score >= 40:
		return "D"
	default:
		return "F"
	}
}
