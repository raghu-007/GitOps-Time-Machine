package analyzer

import (
	"testing"

	"github.com/raghu-007/GitOps-Time-Machine/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCompare_NoChanges(t *testing.T) {
	base := &types.ResourceSnapshot{
		Resources: []types.Resource{
			{Kind: "Deployment", Namespace: "default", Name: "nginx", Labels: map[string]string{"app": "nginx"}},
			{Kind: "Service", Namespace: "default", Name: "nginx-svc"},
		},
	}
	target := &types.ResourceSnapshot{
		Resources: []types.Resource{
			{Kind: "Deployment", Namespace: "default", Name: "nginx", Labels: map[string]string{"app": "nginx"}},
			{Kind: "Service", Namespace: "default", Name: "nginx-svc"},
		},
	}

	report := New().Compare(base, target)

	assert.Equal(t, 0, report.Summary.AddedResources)
	assert.Equal(t, 0, report.Summary.RemovedResources)
	assert.Equal(t, 0, report.Summary.ModifiedResources)
	assert.Equal(t, 2, report.Summary.UnchangedResources)
	assert.False(t, HasDrift(report))
}

func TestCompare_AddedResource(t *testing.T) {
	base := &types.ResourceSnapshot{
		Resources: []types.Resource{
			{Kind: "Deployment", Namespace: "default", Name: "nginx"},
		},
	}
	target := &types.ResourceSnapshot{
		Resources: []types.Resource{
			{Kind: "Deployment", Namespace: "default", Name: "nginx"},
			{Kind: "Service", Namespace: "default", Name: "new-svc"},
		},
	}

	report := New().Compare(base, target)

	assert.Equal(t, 1, report.Summary.AddedResources)
	assert.Equal(t, 0, report.Summary.RemovedResources)
	assert.True(t, HasDrift(report))

	// Verify the added entry
	var added *types.DriftEntry
	for _, e := range report.Entries {
		if e.Type == types.DriftAdded {
			added = &e
			break
		}
	}
	assert.NotNil(t, added)
	assert.Equal(t, "new-svc", added.Resource.Name)
}

func TestCompare_RemovedResource(t *testing.T) {
	base := &types.ResourceSnapshot{
		Resources: []types.Resource{
			{Kind: "Deployment", Namespace: "default", Name: "nginx"},
			{Kind: "Service", Namespace: "default", Name: "old-svc"},
		},
	}
	target := &types.ResourceSnapshot{
		Resources: []types.Resource{
			{Kind: "Deployment", Namespace: "default", Name: "nginx"},
		},
	}

	report := New().Compare(base, target)

	assert.Equal(t, 0, report.Summary.AddedResources)
	assert.Equal(t, 1, report.Summary.RemovedResources)
	assert.True(t, HasDrift(report))
}

func TestCompare_ModifiedResource(t *testing.T) {
	base := &types.ResourceSnapshot{
		Resources: []types.Resource{
			{
				Kind:      "Deployment",
				Namespace: "default",
				Name:      "nginx",
				Labels:    map[string]string{"app": "nginx", "version": "1.0"},
			},
		},
	}
	target := &types.ResourceSnapshot{
		Resources: []types.Resource{
			{
				Kind:      "Deployment",
				Namespace: "default",
				Name:      "nginx",
				Labels:    map[string]string{"app": "nginx", "version": "2.0"},
			},
		},
	}

	report := New().Compare(base, target)

	assert.Equal(t, 1, report.Summary.ModifiedResources)
	assert.True(t, HasDrift(report))
}

func TestCompare_SpecChanges(t *testing.T) {
	base := &types.ResourceSnapshot{
		Resources: []types.Resource{
			{
				Kind:      "Deployment",
				Namespace: "default",
				Name:      "nginx",
				Spec:      map[string]interface{}{"replicas": 3},
			},
		},
	}
	target := &types.ResourceSnapshot{
		Resources: []types.Resource{
			{
				Kind:      "Deployment",
				Namespace: "default",
				Name:      "nginx",
				Spec:      map[string]interface{}{"replicas": 5},
			},
		},
	}

	report := New().Compare(base, target)

	assert.Equal(t, 1, report.Summary.ModifiedResources)
	assert.Len(t, report.Entries[0].FieldDiffs, 1)
	assert.Equal(t, ".spec.replicas", report.Entries[0].FieldDiffs[0].Path)
}

func TestCompare_ClusterScopedResources(t *testing.T) {
	base := &types.ResourceSnapshot{
		Resources: []types.Resource{
			{Kind: "ClusterRole", Name: "admin"},
		},
	}
	target := &types.ResourceSnapshot{
		Resources: []types.Resource{
			{Kind: "ClusterRole", Name: "admin"},
			{Kind: "ClusterRole", Name: "viewer"},
		},
	}

	report := New().Compare(base, target)

	assert.Equal(t, 1, report.Summary.AddedResources)
}

func TestFormatReport_NoDrift(t *testing.T) {
	report := &types.DriftReport{
		Summary: types.DriftSummary{
			TotalResources:    5,
			UnchangedResources: 5,
		},
	}

	output := FormatReport(report)
	assert.Contains(t, output, "No drift detected")
}

func TestFormatReport_WithDrift(t *testing.T) {
	report := &types.DriftReport{
		Summary: types.DriftSummary{
			AddedResources: 1,
		},
		Entries: []types.DriftEntry{
			{
				Type:     types.DriftAdded,
				Resource: types.Resource{Kind: "Service", Namespace: "default", Name: "new-svc"},
			},
		},
	}

	output := FormatReport(report)
	assert.Contains(t, output, "ADDED")
	assert.Contains(t, output, "new-svc")
}

func TestResourceFullName(t *testing.T) {
	r := types.Resource{Kind: "Deployment", Namespace: "prod", Name: "api"}
	assert.Equal(t, "prod/Deployment/api", r.FullName())

	clusterR := types.Resource{Kind: "ClusterRole", Name: "admin"}
	assert.Equal(t, "ClusterRole/admin", clusterR.FullName())
}
