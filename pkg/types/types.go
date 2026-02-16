// Package types defines shared data structures used across GitOps-Time-Machine.
package types

import "time"

// Resource represents a single Kubernetes resource's captured state.
type Resource struct {
	APIVersion string                 `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                 `json:"kind" yaml:"kind"`
	Namespace  string                 `json:"namespace" yaml:"namespace"`
	Name       string                 `json:"name" yaml:"name"`
	Labels     map[string]string      `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string     `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Spec       map[string]interface{} `json:"spec,omitempty" yaml:"spec,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`
	Raw        map[string]interface{} `json:"raw,omitempty" yaml:"-"`
}

// FullName returns namespace/kind/name identifier for the resource.
func (r Resource) FullName() string {
	if r.Namespace == "" {
		return r.Kind + "/" + r.Name
	}
	return r.Namespace + "/" + r.Kind + "/" + r.Name
}

// ResourceSnapshot represents a complete point-in-time capture of cluster state.
type ResourceSnapshot struct {
	Metadata  SnapshotMetadata `json:"metadata" yaml:"metadata"`
	Resources []Resource       `json:"resources" yaml:"resources"`
}

// SnapshotMetadata holds information about when and how a snapshot was taken.
type SnapshotMetadata struct {
	Timestamp     time.Time `json:"timestamp" yaml:"timestamp"`
	ClusterName   string    `json:"clusterName" yaml:"clusterName"`
	Context       string    `json:"context" yaml:"context"`
	ResourceCount int       `json:"resourceCount" yaml:"resourceCount"`
	Namespaces    []string  `json:"namespaces" yaml:"namespaces"`
	CommitHash    string    `json:"commitHash,omitempty" yaml:"commitHash,omitempty"`
}

// DriftReport represents the results of comparing two snapshots.
type DriftReport struct {
	Timestamp time.Time    `json:"timestamp" yaml:"timestamp"`
	BaseRef   string       `json:"baseRef" yaml:"baseRef"`
	TargetRef string       `json:"targetRef" yaml:"targetRef"`
	Summary   DriftSummary `json:"summary" yaml:"summary"`
	Entries   []DriftEntry `json:"entries" yaml:"entries"`
}

// DriftSummary provides a high-level overview of the drift.
type DriftSummary struct {
	TotalResources    int `json:"totalResources" yaml:"totalResources"`
	AddedResources    int `json:"addedResources" yaml:"addedResources"`
	RemovedResources  int `json:"removedResources" yaml:"removedResources"`
	ModifiedResources int `json:"modifiedResources" yaml:"modifiedResources"`
	UnchangedResources int `json:"unchangedResources" yaml:"unchangedResources"`
}

// DriftType indicates the kind of drift detected.
type DriftType string

const (
	DriftAdded    DriftType = "ADDED"
	DriftRemoved  DriftType = "REMOVED"
	DriftModified DriftType = "MODIFIED"
)

// DriftEntry represents a single drift item between two snapshots.
type DriftEntry struct {
	Type       DriftType              `json:"type" yaml:"type"`
	Resource   Resource               `json:"resource" yaml:"resource"`
	FieldDiffs []FieldDiff            `json:"fieldDiffs,omitempty" yaml:"fieldDiffs,omitempty"`
}

// FieldDiff represents a change in a specific field of a resource.
type FieldDiff struct {
	Path     string      `json:"path" yaml:"path"`
	OldValue interface{} `json:"oldValue,omitempty" yaml:"oldValue,omitempty"`
	NewValue interface{} `json:"newValue,omitempty" yaml:"newValue,omitempty"`
}

// HistoryEntry represents a single entry in the snapshot history.
type HistoryEntry struct {
	CommitHash    string    `json:"commitHash" yaml:"commitHash"`
	Timestamp     time.Time `json:"timestamp" yaml:"timestamp"`
	Message       string    `json:"message" yaml:"message"`
	ResourceCount int       `json:"resourceCount" yaml:"resourceCount"`
	Author        string    `json:"author" yaml:"author"`
}
