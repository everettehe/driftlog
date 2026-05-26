package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/you/driftlog/internal/history"
	"github.com/you/driftlog/internal/rollup"
)

func runRollup(historyPath, period string) error {
	entries, err := history.Load(historyPath)
	if err != nil {
		return fmt.Errorf("loading history: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("no history entries found")
		return nil
	}

	var rollupEntries []rollup.Entry
	for _, e := range entries {
		rollupEntries = append(rollupEntries, rollup.Entry{
			At:      e.RunAt,
			Results: e.Results,
		})
	}

	var r rollup.Rollup
	switch period {
	case "week":
		r = rollup.ByWeek(rollupEntries)
	default:
		r = rollup.ByDay(rollupEntries)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PERIOD\tSTART\tTOTAL\tDRIFTED\tMISSING\tUNMANAGED")
	for _, p := range r.Periods {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%d\n",
			p.Label,
			p.Start.Format(time.DateOnly),
			p.Total,
			p.Drifted,
			p.Missing,
			p.Unmanaged,
		)
	}
	return w.Flush()
}
