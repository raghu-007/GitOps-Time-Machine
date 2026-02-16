// Package cmd implements the CLI commands for GitOps-Time-Machine.
package cmd

import (
	"fmt"
	"os"

	"github.com/raghu-007/GitOps-Time-Machine/internal/logger"
	"github.com/raghu-007/GitOps-Time-Machine/internal/printer"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	kubeconfig string
	verbose    bool
	cfg        *config.Config
	version    string
	buildTime  string
)

// SetVersionInfo sets the version info from build-time ldflags.
func SetVersionInfo(v, bt string) {
	version = v
	buildTime = bt
}

// rootCmd represents the base command.
var rootCmd = &cobra.Command{
	Use:   "gitops-time-machine",
	Short: "Infrastructure time-travel & drift detection",
	Long: `GitOps-Time-Machine continuously versions the actual state of live 
infrastructure into a Git repository, enabling time-travel debugging 
and drift analysis.

Capture snapshots, detect drift, and travel back in time to see
exactly what your infrastructure looked like at any point.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override kubeconfig if provided via flag
		if kubeconfig != "" {
			cfg.Kubeconfig = kubeconfig
		}

		// Set log level
		logLevel := cfg.Log.Level
		if verbose {
			logLevel = "debug"
		}
		logger.Init(logLevel, cfg.Log.Format)

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		printer.Banner()
		fmt.Printf("\n  Version:    %s\n", version)
		fmt.Printf("  Build Time: %s\n\n", buildTime)
		fmt.Println("  Run 'gitops-time-machine --help' for usage information.")
		fmt.Println()
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config.yaml)")
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose/debug output")

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GitOps-Time-Machine %s (built %s)\n", version, buildTime)
		},
	})
}

// getConfig returns the loaded config (for use by subcommands).
func getConfig() *config.Config {
	if cfg == nil {
		// Fallback to defaults if PersistentPreRunE hasn't been called
		cfg = config.DefaultConfig()
	}
	return cfg
}

// exitOnError prints an error message and exits.
func exitOnError(err error) {
	printer.Error(err.Error())
	os.Exit(1)
}
