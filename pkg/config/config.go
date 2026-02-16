// Package config handles configuration loading for GitOps-Time-Machine.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for GitOps-Time-Machine.
type Config struct {
	Kubeconfig string          `mapstructure:"kubeconfig"`
	Context    string          `mapstructure:"context"`
	Snapshot   SnapshotConfig  `mapstructure:"snapshot"`
	Git        GitConfig       `mapstructure:"git"`
	Watch      WatchConfig     `mapstructure:"watch"`
	Log        LogConfig       `mapstructure:"log"`
}

// SnapshotConfig configures what resources to capture.
type SnapshotConfig struct {
	OutputDir         string   `mapstructure:"output_dir"`
	ResourceTypes     []string `mapstructure:"resource_types"`
	Namespaces        []string `mapstructure:"namespaces"`
	ExcludeNamespaces []string `mapstructure:"exclude_namespaces"`
	StripFields       []string `mapstructure:"strip_fields"`
}

// GitConfig configures the snapshot Git repository.
type GitConfig struct {
	AuthorName          string `mapstructure:"author_name"`
	AuthorEmail         string `mapstructure:"author_email"`
	CommitMessagePrefix string `mapstructure:"commit_message_prefix"`
	Branch              string `mapstructure:"branch"`
}

// WatchConfig configures scheduled/continuous snapshots.
type WatchConfig struct {
	Schedule          string `mapstructure:"schedule"`
	EnableWatchEvents bool   `mapstructure:"enable_watch_events"`
}

// LogConfig configures logging.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Kubeconfig: defaultKubeconfig(),
		Snapshot: SnapshotConfig{
			OutputDir: "./infra-snapshots",
			ResourceTypes: []string{
				"deployments", "services", "configmaps", "secrets",
				"ingresses", "statefulsets", "daemonsets", "cronjobs",
				"persistentvolumeclaims", "networkpolicies",
				"serviceaccounts", "roles", "rolebindings",
			},
			ExcludeNamespaces: []string{
				"kube-system", "kube-public", "kube-node-lease",
			},
			StripFields: []string{
				".metadata.managedFields",
				".metadata.resourceVersion",
				".metadata.uid",
				".metadata.generation",
				".status",
			},
		},
		Git: GitConfig{
			AuthorName:          "GitOps-Time-Machine",
			AuthorEmail:         "gitops-tm@automated",
			CommitMessagePrefix: "[snapshot]",
			Branch:              "main",
		},
		Watch: WatchConfig{
			Schedule: "*/5 * * * *",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// Load reads the configuration from file, environment, and flags.
func Load(cfgFile string) (*Config, error) {
	cfg := DefaultConfig()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.gitops-time-machine")
		viper.AddConfigPath("/etc/gitops-time-machine")
	}

	// Environment variable support: GTM_KUBECONFIG, GTM_SNAPSHOT_OUTPUT_DIR, etc.
	viper.SetEnvPrefix("GTM")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK — use defaults
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	return cfg, nil
}

// defaultKubeconfig returns the default kubeconfig path.
func defaultKubeconfig() string {
	if env := os.Getenv("KUBECONFIG"); env != "" {
		return env
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".kube", "config")
}
