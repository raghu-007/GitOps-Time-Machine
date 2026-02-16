package cmd

import (
	"fmt"

	"github.com/raghu-007/GitOps-Time-Machine/internal/printer"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/versioner"
	"github.com/spf13/cobra"
)

var historyLimit int

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "List all infrastructure snapshots",
	Long: `Shows a chronological list of all committed snapshots, 
including timestamps, commit hashes, and resource counts.`,
	Example: `  # Show last 10 snapshots
  gitops-time-machine history --limit 10
  
  # Show all snapshots
  gitops-time-machine history`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()

		ver, err := versioner.New(cfg.Snapshot.OutputDir, &cfg.Git)
		if err != nil {
			return fmt.Errorf("failed to initialize versioner: %w", err)
		}

		entries, err := ver.History(historyLimit)
		if err != nil {
			return fmt.Errorf("failed to get history: %w", err)
		}

		commitCount, _ := ver.GetCommitCount()

		printer.Banner()

		if historyLimit > 0 && commitCount > historyLimit {
			printer.Info(fmt.Sprintf("Showing last %d of %d total snapshots", historyLimit, commitCount))
		}

		printer.HistoryTable(entries)

		return nil
	},
}

func init() {
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 20, "maximum number of entries to show (0 = all)")

	rootCmd.AddCommand(historyCmd)
}
