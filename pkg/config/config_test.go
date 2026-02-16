package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "./infra-snapshots", cfg.Snapshot.OutputDir)
	assert.Equal(t, "main", cfg.Git.Branch)
	assert.Equal(t, "GitOps-Time-Machine", cfg.Git.AuthorName)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "text", cfg.Log.Format)
	assert.Equal(t, "*/5 * * * *", cfg.Watch.Schedule)
	assert.Contains(t, cfg.Snapshot.ResourceTypes, "deployments")
	assert.Contains(t, cfg.Snapshot.ExcludeNamespaces, "kube-system")
}

func TestDefaultConfig_StripFields(t *testing.T) {
	cfg := DefaultConfig()

	assert.Contains(t, cfg.Snapshot.StripFields, ".metadata.managedFields")
	assert.Contains(t, cfg.Snapshot.StripFields, ".status")
}

func TestLoad_MissingConfigFile(t *testing.T) {
	// Loading with a missing config file should return defaults
	cfg, err := Load("")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "./infra-snapshots", cfg.Snapshot.OutputDir)
}

func TestLoad_InvalidConfigFile(t *testing.T) {
	// Loading with an invalid config file should return error
	_, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err)
}

func TestDefaultKubeconfig_EnvVar(t *testing.T) {
	// Test that KUBECONFIG env var is respected
	original := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", original)

	os.Setenv("KUBECONFIG", "/custom/kubeconfig")
	path := defaultKubeconfig()
	assert.Equal(t, "/custom/kubeconfig", path)
}

func TestDefaultKubeconfig_Default(t *testing.T) {
	// Test default path when no env var is set
	original := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", original)

	os.Unsetenv("KUBECONFIG")
	path := defaultKubeconfig()
	assert.Contains(t, path, ".kube")
	assert.Contains(t, path, "config")
}
