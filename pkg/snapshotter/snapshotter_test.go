package snapshotter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/raghu-007/GitOps-Time-Machine/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()

	snap := New(tmpDir)

	original := &types.ResourceSnapshot{
		Metadata: types.SnapshotMetadata{
			Timestamp:     time.Now().UTC().Truncate(time.Second),
			ClusterName:   "test-cluster",
			Context:       "test-context",
			ResourceCount: 2,
			Namespaces:    []string{"default", "monitoring"},
		},
		Resources: []types.Resource{
			{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Namespace:  "default",
				Name:       "nginx",
				Labels:     map[string]string{"app": "nginx"},
			},
			{
				APIVersion: "v1",
				Kind:       "Service",
				Namespace:  "monitoring",
				Name:       "prometheus",
			},
		},
	}

	// Write snapshot
	err := snap.Write(original)
	require.NoError(t, err)

	// Verify metadata file exists
	metadataPath := filepath.Join(tmpDir, "_metadata.yaml")
	assert.FileExists(t, metadataPath)

	// Verify resource files exist
	deploymentPath := filepath.Join(tmpDir, "default", "deployment", "nginx.yaml")
	assert.FileExists(t, deploymentPath)

	servicePath := filepath.Join(tmpDir, "monitoring", "service", "prometheus.yaml")
	assert.FileExists(t, servicePath)

	// Read snapshot back
	readSnap, err := snap.Read()
	require.NoError(t, err)

	assert.Equal(t, "test-cluster", readSnap.Metadata.ClusterName)
	assert.Equal(t, 2, readSnap.Metadata.ResourceCount)
}

func TestWriteClusterScopedResources(t *testing.T) {
	tmpDir := t.TempDir()

	snap := New(tmpDir)

	snapshot := &types.ResourceSnapshot{
		Metadata: types.SnapshotMetadata{
			Timestamp:     time.Now().UTC(),
			ResourceCount: 1,
		},
		Resources: []types.Resource{
			{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRole",
				Name:       "admin",
			},
		},
	}

	err := snap.Write(snapshot)
	require.NoError(t, err)

	// Cluster-scoped resources go under _cluster/
	clusterRolePath := filepath.Join(tmpDir, "_cluster", "clusterrole", "admin.yaml")
	assert.FileExists(t, clusterRolePath)
}

func TestCleanDirectory_PreservesGit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main"), 0644))

	// Create some other content
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "old-file.yaml"), []byte("old"), 0644))

	snap := New(tmpDir)
	err := snap.cleanDirectory()
	require.NoError(t, err)

	// .git should still exist
	assert.DirExists(t, gitDir)

	// Old file should be gone
	assert.NoFileExists(t, filepath.Join(tmpDir, "old-file.yaml"))
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple-name", "simple-name"},
		{"name:with:colons", "name_with_colons"},
		{"name/with/slashes", "name_with_slashes"},
		{"name<with>angles", "name_with_angles"},
	}

	for _, tc := range tests {
		result := sanitizeFilename(tc.input)
		assert.Equal(t, tc.expected, result, "sanitizeFilename(%q)", tc.input)
	}
}
