package cmd

import (
	"context"
	"fmt"

	"github.com/raghu-007/GitOps-Time-Machine/internal/printer"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/analyzer"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/collector"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/snapshotter"
	"github.com/spf13/cobra"
)

var driftCmd = &cobra.Command{
	Use:   "drift",
	Short: "Detect drift between live state and last snapshot",
	Long: `Captures the current live infrastructure state and compares it 
against the last committed snapshot. Shows any resources that have 
been added, removed, or modified since the last snapshot.

This is useful for detecting manual changes, unauthorized 
modifications, or configuration drift.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()

		printer.Banner()
		printer.Info("Checking for infrastructure drift...")

		// Read the last committed snapshot
		snap := snapshotter.New(cfg.Snapshot.OutputDir)
		lastSnapshot, err := snap.Read()
		if err != nil {
			return fmt.Errorf("failed to read last snapshot (run 'snapshot' first): %w", err)
		}

		// Collect current live state
		coll, err := collector.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to create collector: %w", err)
		}

		ctx := context.Background()
		liveSnapshot, err := coll.Collect(ctx)
		if err != nil {
			return fmt.Errorf("failed to collect live state: %w", err)
		}

		// Compare
		report := analyzer.New().Compare(lastSnapshot, liveSnapshot)

		// Print results
		printer.DriftSummary(report)

		if analyzer.HasDrift(report) {
			printer.Info("Drift detected! Review the changes above.")
			printer.Info("Run 'gitops-time-machine snapshot' to capture the current state.")
		} else {
			printer.Success("No drift detected — infrastructure matches the last snapshot.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(driftCmd)
}
