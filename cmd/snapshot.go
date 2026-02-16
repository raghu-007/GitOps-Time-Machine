package cmd

import (
	"context"
	"fmt"

	"github.com/raghu-007/GitOps-Time-Machine/internal/printer"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/collector"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/snapshotter"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/versioner"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Capture a point-in-time snapshot of infrastructure state",
	Long: `Connects to the configured Kubernetes cluster, captures the current 
state of all configured resources, writes them as organized YAML files, 
and commits the snapshot to the Git repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()

		printer.Banner()
		printer.Info("Starting infrastructure snapshot...")

		// Create collector
		coll, err := collector.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to create collector: %w", err)
		}

		// Collect resources
		ctx := context.Background()
		snapshot, err := coll.Collect(ctx)
		if err != nil {
			return fmt.Errorf("failed to collect resources: %w", err)
		}

		// Write to disk
		snap := snapshotter.New(cfg.Snapshot.OutputDir)
		if err := snap.Write(snapshot); err != nil {
			return fmt.Errorf("failed to write snapshot: %w", err)
		}

		// Commit to Git
		ver, err := versioner.New(cfg.Snapshot.OutputDir, &cfg.Git)
		if err != nil {
			return fmt.Errorf("failed to initialize versioner: %w", err)
		}

		commitHash, err := ver.Commit(&snapshot.Metadata)
		if err != nil {
			return fmt.Errorf("failed to commit snapshot: %w", err)
		}

		snapshot.Metadata.CommitHash = commitHash

		// Print summary
		printer.SnapshotSummary(&snapshot.Metadata)
		printer.Success("Snapshot captured and committed successfully!")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
}
