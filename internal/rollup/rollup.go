// Package rollup aggregates drift results into a concise period-based summary.
package rollup

import (
	"fmt"
	"sort"
	"time"

	"github.com/you/driftlog/internal/diff"
)

// Period groups results by a named time bucket.
type Period struct {
	Label     string
	Start     time.Time
	End       time.Time
	Total     int
	Drifted   int
	Missing   int
	Unmanaged int
}

// Rollup holds aggregated periods in chronological order.
type Rollup struct {
	Periods []Period
}

// Entry is a timestamped snapshot of drift results.
type Entry struct {
	At      time.Time
	Results []diff.Result
}

// ByDay aggregates entries into daily periods.
func ByDay(entries []Entry) Rollup {
	return aggregate(entries, func(t time.Time) string {
		return t.UTC().Format("2006-01-02")
	}, 24*time.Hour)
}

// ByWeek aggregates entries into ISO-week periods.
func ByWeek(entries []Entry) Rollup {
	return aggregate(entries, func(t time.Time) string {
		y, w := t.UTC().ISOWeek()
		return fmt.Sprintf("%d-W%02d", y, w)
	}, 7*24*time.Hour)
}

func aggregate(entries []Entry, label func(time.Time) string, width time.Duration) Rollup {
	type bucket struct {
		period  Period
		eariest time.Time
	}
	buckets := map[string]*bucket{}

	for _, e := range entries {
		k := label(e.At)
		b, ok := buckets[k]
		if !ok {
			b = &bucket{period: Period{Label: k}, eariest: e.At}
			buckets[k] = b
		}
		if e.At.Before(b.eariest) {
			b.eariest = e.At
		}
		for _, r := range e.Results {
			b.period.Total++
			switch r.Status {
			case diff.StatusDrifted:
				b.period.Drifted++
			case diff.StatusOnlyInState:
				b.period.Missing++
			case diff.StatusOnlyInCloud:
				b.period.Unmanaged++
			}
		}
	}

	var periods []Period
	for _, b := range buckets {
		b.period.Start = b.eariest.UTC().Truncate(width)
		b.period.End = b.period.Start.Add(width)
		periods = append(periods, b.period)
	}
	sort.Slice(periods, func(i, j int) bool {
		return periods[i].Label < periods[j].Label
	})
	return Rollup{Periods: periods}
}
