// Package printer provides terminal output formatting utilities.
package printer

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/types"
)

var (
	// Color helpers
	green  = color.New(color.FgGreen, color.Bold).SprintFunc()
	red    = color.New(color.FgRed, color.Bold).SprintFunc()
	yellow = color.New(color.FgYellow, color.Bold).SprintFunc()
	cyan   = color.New(color.FgCyan, color.Bold).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
	dim    = color.New(color.Faint).SprintFunc()
)

// Banner prints the application banner.
func Banner() {
	banner := `
  ╔══════════════════════════════════════════════╗
  ║       GitOps-Time-Machine  ⏰ → 🔀 → 📦      ║
  ║   Infrastructure Time-Travel & Drift Detect  ║
  ╚══════════════════════════════════════════════╝`
	fmt.Println(cyan(banner))
}

// SnapshotSummary prints a summary of a completed snapshot.
func SnapshotSummary(metadata *types.SnapshotMetadata) {
	fmt.Println()
	fmt.Println(bold("📸 Snapshot Captured"))
	fmt.Println(strings.Repeat("─", 45))
	fmt.Printf("  ⏰  Time:       %s\n", metadata.Timestamp.Format("2006-01-02 15:04:05 UTC"))
	fmt.Printf("  🏗️  Cluster:    %s\n", metadata.ClusterName)
	fmt.Printf("  📦  Resources:  %s\n", green(fmt.Sprintf("%d", metadata.ResourceCount)))
	fmt.Printf("  🗂️  Namespaces: %s\n", cyan(fmt.Sprintf("%d", len(metadata.Namespaces))))
	if metadata.CommitHash != "" {
		fmt.Printf("  🔗  Commit:     %s\n", dim(metadata.CommitHash[:8]))
	}
	fmt.Println()
}

// HistoryTable prints the snapshot history as a formatted table.
func HistoryTable(entries []types.HistoryEntry) {
	if len(entries) == 0 {
		fmt.Println(yellow("No snapshots found."))
		return
	}

	fmt.Println()
	fmt.Println(bold("📜 Snapshot History"))
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Commit", "Timestamp", "Message"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(true)

	for i, entry := range entries {
		hash := entry.CommitHash
		if len(hash) > 8 {
			hash = hash[:8]
		}
		msg := entry.Message
		if len(msg) > 50 {
			msg = msg[:50] + "..."
		}
		table.Append([]string{
			fmt.Sprintf("%d", i+1),
			hash,
			entry.Timestamp.Format("2006-01-02 15:04:05"),
			msg,
		})
	}

	table.Render()
	fmt.Println()
}

// DriftSummary prints a summary of drift analysis.
func DriftSummary(report *types.DriftReport) {
	fmt.Println()
	fmt.Println(bold("🔍 Drift Analysis"))
	fmt.Println(strings.Repeat("─", 45))

	if len(report.Entries) == 0 {
		fmt.Println(green("  ✅ No drift detected — infrastructure matches!"))
		fmt.Println()
		return
	}

	fmt.Printf("  Added:     %s\n", green(fmt.Sprintf("+%d", report.Summary.AddedResources)))
	fmt.Printf("  Removed:   %s\n", red(fmt.Sprintf("-%d", report.Summary.RemovedResources)))
	fmt.Printf("  Modified:  %s\n", yellow(fmt.Sprintf("~%d", report.Summary.ModifiedResources)))
	fmt.Printf("  Unchanged: %s\n", dim(fmt.Sprintf("%d", report.Summary.UnchangedResources)))
	fmt.Println()

	for _, entry := range report.Entries {
		switch entry.Type {
		case types.DriftAdded:
			fmt.Printf("  %s %s\n", green("[+]"), entry.Resource.FullName())
		case types.DriftRemoved:
			fmt.Printf("  %s %s\n", red("[-]"), entry.Resource.FullName())
		case types.DriftModified:
			fmt.Printf("  %s %s\n", yellow("[~]"), entry.Resource.FullName())
			for _, diff := range entry.FieldDiffs {
				fmt.Printf("      %s %s\n", dim("•"), diff.Path)
				if diff.OldValue != nil {
					fmt.Printf("        %s %v\n", red("-"), diff.OldValue)
				}
				if diff.NewValue != nil {
					fmt.Printf("        %s %v\n", green("+"), diff.NewValue)
				}
			}
		}
	}
	fmt.Println()
}

// Success prints a success message.
func Success(msg string) {
	fmt.Printf("%s %s\n", green("✓"), msg)
}

// Error prints an error message.
func Error(msg string) {
	fmt.Printf("%s %s\n", red("✗"), msg)
}

// Info prints an info message.
func Info(msg string) {
	fmt.Printf("%s %s\n", cyan("ℹ"), msg)
}
