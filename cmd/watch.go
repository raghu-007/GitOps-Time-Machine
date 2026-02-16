package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/raghu-007/GitOps-Time-Machine/internal/printer"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/collector"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/scheduler"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/snapshotter"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/versioner"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var watchSchedule string

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Continuously capture snapshots on a schedule",
	Long: `Starts a background process that takes infrastructure snapshots 
at regular intervals using a cron schedule. Runs until interrupted.

Default schedule: every 5 minutes (configured in config file or via --schedule flag).`,
	Example: `  # Watch with default schedule (every 5 minutes)
  gitops-time-machine watch
  
  # Watch every minute
  gitops-time-machine watch --schedule "* * * * *"
  
  # Watch every hour
  gitops-time-machine watch --schedule "0 * * * *"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()

		schedule := cfg.Watch.Schedule
		if watchSchedule != "" {
			schedule = watchSchedule
		}

		printer.Banner()
		printer.Info(fmt.Sprintf("Starting continuous watch with schedule: %s", schedule))
		printer.Info("Press Ctrl+C to stop.")
		fmt.Println()

		// Create the snapshot function
		snapshotFn := func(ctx context.Context) error {
			coll, err := collector.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to create collector: %w", err)
			}

			snapshot, err := coll.Collect(ctx)
			if err != nil {
				return fmt.Errorf("failed to collect resources: %w", err)
			}

			snap := snapshotter.New(cfg.Snapshot.OutputDir)
			if err := snap.Write(snapshot); err != nil {
				return fmt.Errorf("failed to write snapshot: %w", err)
			}

			ver, err := versioner.New(cfg.Snapshot.OutputDir, &cfg.Git)
			if err != nil {
				return fmt.Errorf("failed to initialize versioner: %w", err)
			}

			commitHash, err := ver.Commit(&snapshot.Metadata)
			if err != nil {
				return fmt.Errorf("failed to commit: %w", err)
			}

			if commitHash != "" {
				snapshot.Metadata.CommitHash = commitHash
				printer.SnapshotSummary(&snapshot.Metadata)
			} else {
				printer.Info("No changes detected, skipping commit.")
			}

			return nil
		}

		// Create scheduler
		sched, err := scheduler.New(schedule, snapshotFn)
		if err != nil {
			return fmt.Errorf("failed to create scheduler: %w", err)
		}

		// Handle graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigCh
			log.Info("received shutdown signal")
			cancel()
		}()

		// Take an initial snapshot immediately
		printer.Info("Taking initial snapshot...")
		if err := snapshotFn(ctx); err != nil {
			log.WithError(err).Warn("initial snapshot failed")
		}

		// Start the scheduler (blocks until context is cancelled)
		return sched.Start(ctx)
	},
}

func init() {
	watchCmd.Flags().StringVar(&watchSchedule, "schedule", "", "cron schedule (overrides config)")

	rootCmd.AddCommand(watchCmd)
}
