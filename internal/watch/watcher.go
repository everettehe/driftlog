// Package watch provides continuous drift monitoring by running drift
// detection on a configurable schedule and emitting results to registered
// handlers.
package watch

import (
	"context"
	"log"
	"time"

	"github.com/yourorg/driftlog/internal/diff"
	"github.com/yourorg/driftlog/internal/schedule"
)

// RunFunc is the signature of a function that performs a single drift-detection
// pass and returns the results.
type RunFunc func(ctx context.Context) ([]diff.Result, error)

// HandlerFunc is called after every successful detection pass with the
// collected results. Handlers are invoked synchronously; long-running work
// should be dispatched to a goroutine inside the handler.
type HandlerFunc func(results []diff.Result)

// Watcher runs a RunFunc on a schedule and forwards results to one or more
// HandlerFuncs.
type Watcher struct {
	sched    *schedule.Schedule
	runFn    RunFunc
	handlers []HandlerFunc
	logger   *log.Logger
}

// New creates a Watcher that will call runFn according to sched and forward
// results to each handler in order.
func New(sched *schedule.Schedule, runFn RunFunc, logger *log.Logger, handlers ...HandlerFunc) *Watcher {
	if logger == nil {
		logger = log.Default()
	}
	return &Watcher{
		sched:    sched,
		runFn:    runFn,
		handlers: handlers,
		logger:   logger,
	}
}

// Start blocks and runs the detection loop until ctx is cancelled. It executes
// an initial pass immediately, then waits for the next scheduled tick before
// running again.
func (w *Watcher) Start(ctx context.Context) error {
	w.logger.Println("[watch] starting drift watcher")

	// Run once immediately so the operator gets feedback right away.
	w.tick(ctx)

	for {
		next := w.sched.NextRun(time.Now())
		waitFor := time.Until(next)
		w.logger.Printf("[watch] next run in %s (at %s)", waitFor.Round(time.Second), next.Format(time.RFC3339))

		select {
		case <-ctx.Done():
			w.logger.Println("[watch] context cancelled, stopping watcher")
			return ctx.Err()
		case <-time.After(waitFor):
			w.tick(ctx)
		}
	}
}

// tick performs a single detection pass and dispatches results to handlers.
func (w *Watcher) tick(ctx context.Context) {
	w.logger.Println("[watch] running drift detection")

	results, err := w.runFn(ctx)
	if err != nil {
		w.logger.Printf("[watch] detection error: %v", err)
		return
	}

	w.logger.Printf("[watch] detection complete, %d resource(s) evaluated", len(results))

	for _, h := range w.handlers {
		h(results)
	}
}
