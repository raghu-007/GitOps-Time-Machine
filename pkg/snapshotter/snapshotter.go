// Package snapshotter serializes resource snapshots to disk as organized YAML files.
package snapshotter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/raghu-007/GitOps-Time-Machine/pkg/types"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Snapshotter writes resource snapshots to disk in an organized directory structure.
type Snapshotter struct {
	outputDir string
}

// New creates a new Snapshotter that writes to the given directory.
func New(outputDir string) *Snapshotter {
	return &Snapshotter{outputDir: outputDir}
}

// Write persists a ResourceSnapshot to disk.
//
// Directory structure:
//
//	<outputDir>/
//	  _metadata.yaml
//	  <namespace>/
//	    <kind>/
//	      <name>.yaml
//	  _cluster/
//	    <kind>/
//	      <name>.yaml
func (s *Snapshotter) Write(snapshot *types.ResourceSnapshot) error {
	log.WithField("outputDir", s.outputDir).Info("writing snapshot to disk")

	// Clean the output directory (except .git)
	if err := s.cleanDirectory(); err != nil {
		return fmt.Errorf("failed to clean output directory: %w", err)
	}

	// Write metadata
	if err := s.writeMetadata(snapshot); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Write each resource
	for _, resource := range snapshot.Resources {
		if err := s.writeResource(resource); err != nil {
			log.WithError(err).WithField("resource", resource.FullName()).Warn("failed to write resource")
			continue
		}
	}

	log.WithField("resources", len(snapshot.Resources)).Info("snapshot written to disk")
	return nil
}

// Read loads a snapshot from the disk directory structure.
func (s *Snapshotter) Read() (*types.ResourceSnapshot, error) {
	metadataPath := filepath.Join(s.outputDir, "_metadata.yaml")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	snapshot := &types.ResourceSnapshot{}
	if err := yaml.Unmarshal(data, &snapshot.Metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Walk the directory and read all resource files
	err = filepath.Walk(s.outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || info.Name() == "_metadata.yaml" || !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		resData, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var resource types.Resource
		if err := yaml.Unmarshal(resData, &resource); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		snapshot.Resources = append(snapshot.Resources, resource)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}

	snapshot.Metadata.ResourceCount = len(snapshot.Resources)
	return snapshot, nil
}

// writeMetadata writes the snapshot metadata file.
func (s *Snapshotter) writeMetadata(snapshot *types.ResourceSnapshot) error {
	data, err := yaml.Marshal(snapshot.Metadata)
	if err != nil {
		return err
	}

	metadataPath := filepath.Join(s.outputDir, "_metadata.yaml")
	return os.WriteFile(metadataPath, data, 0644)
}

// writeResource writes a single resource to its appropriate file path.
func (s *Snapshotter) writeResource(resource types.Resource) error {
	// Determine directory: namespace-scoped vs cluster-scoped
	var dir string
	if resource.Namespace == "" {
		dir = filepath.Join(s.outputDir, "_cluster", strings.ToLower(resource.Kind))
	} else {
		dir = filepath.Join(s.outputDir, resource.Namespace, strings.ToLower(resource.Kind))
	}

	// Create directory structure
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Sanitize name for filename
	filename := sanitizeFilename(resource.Name) + ".yaml"
	filePath := filepath.Join(dir, filename)

	// Marshal to YAML (use Raw if available for fidelity, otherwise struct)
	var data []byte
	var err error
	if resource.Raw != nil {
		data, err = yaml.Marshal(resource.Raw)
	} else {
		data, err = yaml.Marshal(resource)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal resource: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// cleanDirectory removes all content except .git directory.
func (s *Snapshotter) cleanDirectory() error {
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(s.outputDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}
		path := filepath.Join(s.outputDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	return nil
}

// sanitizeFilename replaces characters that are invalid in filenames.
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}
