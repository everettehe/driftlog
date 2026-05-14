package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/user/driftlog/internal/aws"
	"github.com/user/driftlog/internal/diff"
	"github.com/user/driftlog/internal/report"
	"github.com/user/driftlog/internal/tfstate"
)

func main() {
	statePath := flag.String("state", "terraform.tfstate", "path to terraform state file")
	region := flag.String("region", "us-east-1", "AWS region to inspect")
	flag.Parse()

	// Parse Terraform state
	state, err := tfstate.ParseFile(*statePath)
	if err != nil {
		log.Fatalf("failed to parse state file: %v", err)
	}

	// Fetch live AWS resources
	fetcher, err := aws.NewFetcher(*region)
	if err != nil {
		log.Fatalf("failed to create AWS fetcher: %v", err)
	}

	liveResources, err := fetcher.FetchAll()
	if err != nil {
		log.Fatalf("failed to fetch live resources: %v", err)
	}

	// Compare state vs live
	results := diff.Compare(state.Resources, liveResources)

	// Write report
	if err := report.WriteText(os.Stdout, results); err != nil {
		log.Fatalf("failed to write report: %v", err)
	}

	// Exit with non-zero code if drift detected
	for _, r := range results {
		if r.Status != diff.StatusMatch {
			fmt.Fprintln(os.Stderr, "drift detected")
			os.Exit(1)
		}
	}
}
