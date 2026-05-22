package metrics

import (
	"sync"
	"time"

	"github.com/driftlog/internal/diff"
)

// RunMetrics captures statistics from a single drift detection run.
type RunMetrics struct {
	mu sync.Mutex

	StartedAt    time.Time
	FinishedAt   time.Time
	TotalChecked int
	Drifted      int
	Missing      int
	Extra        int
	Errors       int
}

// New returns a RunMetrics with StartedAt set to now.
func New() *RunMetrics {
	return &RunMetrics{StartedAt: time.Now()}
}

// Record walks a slice of DriftResult values and tallies counts.
func (m *RunMetrics) Record(results []diff.DriftResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalChecked = len(results)
	for _, r := range results {
		switch r.Status {
		case diff.StatusDrifted:
			m.Drifted++
		case diff.StatusMissing:
			m.Missing++
		case diff.StatusExtra:
			m.Extra++
		}
	}
}

// Finish marks the run as complete.
func (m *RunMetrics) Finish() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FinishedAt = time.Now()
}

// Duration returns the elapsed time between start and finish.
// If Finish has not been called, it returns the time since start.
func (m *RunMetrics) Duration() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.FinishedAt.IsZero() {
		return time.Since(m.StartedAt)
	}
	return m.FinishedAt.Sub(m.StartedAt)
}

// Summary returns a human-readable one-line summary of the run.
func (m *RunMetrics) Summary() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fmt.Sprintf(
		"checked=%d drifted=%d missing=%d extra=%d errors=%d duration=%s",
		m.TotalChecked, m.Drifted, m.Missing, m.Extra, m.Errors,
		m.FinishedAt.Sub(m.StartedAt).Round(time.Millisecond),
	)
}
