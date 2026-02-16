package cmd

import (
	"fmt"
	"time"

	"github.com/raghu-007/GitOps-Time-Machine/internal/printer"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/analyzer"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/snapshotter"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/timetravel"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/types"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/versioner"
	"github.com/spf13/cobra"
)

var (
	diffFrom   string
	diffTo     string
	diffCommit string
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show differences between two snapshots",
	Long: `Compare infrastructure state between two points in time or 
two specific commits. Shows added, removed, and modified resources 
with field-level detail.`,
	Example: `  # Compare by timestamps
  gitops-time-machine diff --from "2024-01-01T00:00:00Z" --to "2024-01-02T00:00:00Z"
  
  # Compare current state with a specific commit
  gitops-time-machine diff --commit abc1234`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()

		printer.Banner()
		printer.Info("Analyzing infrastructure differences...")

		ver, err := versioner.New(cfg.Snapshot.OutputDir, &cfg.Git)
		if err != nil {
			return fmt.Errorf("failed to initialize versioner: %w", err)
		}

		snap := snapshotter.New(cfg.Snapshot.OutputDir)
		tt := timetravel.New(ver, snap, cfg.Snapshot.OutputDir)

		var fromSnapshot *types.ResourceSnapshot
		var toSnapshot *types.ResourceSnapshot

		if diffCommit != "" {
			// Compare specific commit with latest
			fromSnap, err := tt.SnapshotByCommit(diffCommit)
			if err != nil {
				return fmt.Errorf("failed to get snapshot for commit %s: %w", diffCommit, err)
			}
			fromSnapshot = fromSnap

			// Get latest snapshot
			toSnap, err := snap.Read()
			if err != nil {
				return fmt.Errorf("failed to read current snapshot: %w", err)
			}
			toSnapshot = toSnap
		} else if diffFrom != "" && diffTo != "" {
			fromTime, err := time.Parse(time.RFC3339, diffFrom)
			if err != nil {
				return fmt.Errorf("invalid --from time format (use RFC3339): %w", err)
			}
			toTime, err := time.Parse(time.RFC3339, diffTo)
			if err != nil {
				return fmt.Errorf("invalid --to time format (use RFC3339): %w", err)
			}

			fromSnap, toSnap, err := tt.CompareTimeRange(fromTime, toTime)
			if err != nil {
				return fmt.Errorf("failed to compare time range: %w", err)
			}
			fromSnapshot = fromSnap
			toSnapshot = toSnap
		} else {
			return fmt.Errorf("specify either --commit or both --from and --to")
		}

		// Run drift analysis
		report := analyzer.New().Compare(fromSnapshot, toSnapshot)
		printer.DriftSummary(report)

		return nil
	},
}

func init() {
	diffCmd.Flags().StringVar(&diffFrom, "from", "", "start time (RFC3339 format)")
	diffCmd.Flags().StringVar(&diffTo, "to", "", "end time (RFC3339 format)")
	diffCmd.Flags().StringVar(&diffCommit, "commit", "", "compare with specific commit hash")

	rootCmd.AddCommand(diffCmd)
}
