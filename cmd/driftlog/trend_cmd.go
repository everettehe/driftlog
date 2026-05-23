package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/user/driftlog/internal/history"
	"github.com/user/driftlog/internal/trend"
)

// runTrend loads drift history and prints a trend analysis to stdout.
// Usage: driftlog trend [window-days] [history-file]
func runTrend(args []string) {
	window := 30
	historyPath := "drift-history.json"

	if len(args) >= 1 {
		w, err := strconv.Atoi(args[0])
		if err != nil || w <= 0 {
			fmt.Fprintf(os.Stderr, "trend: invalid window %q, must be a positive integer\n", args[0])
			os.Exit(1)
		}
		window = w
	}
	if len(args) >= 2 {
		historyPath = args[1]
	}

	entries, err := history.Load(historyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "trend: failed to load history: %v\n", err)
		os.Exit(1)
	}
	if len(entries) == 0 {
		fmt.Println("No drift history found.")
		return
	}

	summary := trend.Analyze(entries, window)
	for _, line := range trend.Lines(summary) {
		fmt.Println(line)
	}
}
