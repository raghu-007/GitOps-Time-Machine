// Package analyzer provides drift detection between infrastructure snapshots.
package analyzer

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/raghu-007/GitOps-Time-Machine/pkg/types"
	log "github.com/sirupsen/logrus"
)

// Analyzer compares infrastructure snapshots and detects drift.
type Analyzer struct{}

// New creates a new Analyzer.
func New() *Analyzer {
	return &Analyzer{}
}

// Compare takes two snapshots and produces a DriftReport.
func (a *Analyzer) Compare(base, target *types.ResourceSnapshot) *types.DriftReport {
	report := &types.DriftReport{
		Timestamp: time.Now().UTC(),
		BaseRef:   base.Metadata.CommitHash,
		TargetRef: target.Metadata.CommitHash,
	}

	// Index resources by their full name for efficient lookup
	baseIndex := indexResources(base.Resources)
	targetIndex := indexResources(target.Resources)

	// Find removed resources (in base but not in target)
	for name, baseRes := range baseIndex {
		if _, exists := targetIndex[name]; !exists {
			report.Entries = append(report.Entries, types.DriftEntry{
				Type:     types.DriftRemoved,
				Resource: baseRes,
			})
		}
	}

	// Find added resources (in target but not in base)
	for name, targetRes := range targetIndex {
		if _, exists := baseIndex[name]; !exists {
			report.Entries = append(report.Entries, types.DriftEntry{
				Type:     types.DriftAdded,
				Resource: targetRes,
			})
		}
	}

	// Find modified resources (in both, but different)
	for name, baseRes := range baseIndex {
		if targetRes, exists := targetIndex[name]; exists {
			diffs := compareResources(baseRes, targetRes)
			if len(diffs) > 0 {
				report.Entries = append(report.Entries, types.DriftEntry{
					Type:       types.DriftModified,
					Resource:   targetRes,
					FieldDiffs: diffs,
				})
			}
		}
	}

	// Sort entries for deterministic output
	sort.Slice(report.Entries, func(i, j int) bool {
		if report.Entries[i].Type != report.Entries[j].Type {
			return report.Entries[i].Type < report.Entries[j].Type
		}
		return report.Entries[i].Resource.FullName() < report.Entries[j].Resource.FullName()
	})

	// Build summary
	report.Summary = types.DriftSummary{
		TotalResources: len(targetIndex),
	}
	for _, entry := range report.Entries {
		switch entry.Type {
		case types.DriftAdded:
			report.Summary.AddedResources++
		case types.DriftRemoved:
			report.Summary.RemovedResources++
		case types.DriftModified:
			report.Summary.ModifiedResources++
		}
	}
	report.Summary.UnchangedResources = len(baseIndex) - report.Summary.RemovedResources - report.Summary.ModifiedResources

	log.WithFields(log.Fields{
		"added":    report.Summary.AddedResources,
		"removed":  report.Summary.RemovedResources,
		"modified": report.Summary.ModifiedResources,
	}).Info("drift analysis completed")

	return report
}

// HasDrift returns true if the report contains any drift entries.
func HasDrift(report *types.DriftReport) bool {
	return len(report.Entries) > 0
}

// FormatReport creates a human-readable string from a DriftReport.
func FormatReport(report *types.DriftReport) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Drift Report — %s\n", report.Timestamp.Format(time.RFC3339)))
	sb.WriteString(strings.Repeat("═", 60) + "\n\n")

	sb.WriteString(fmt.Sprintf("  Total Resources: %d\n", report.Summary.TotalResources))
	sb.WriteString(fmt.Sprintf("  Added:           %d\n", report.Summary.AddedResources))
	sb.WriteString(fmt.Sprintf("  Removed:         %d\n", report.Summary.RemovedResources))
	sb.WriteString(fmt.Sprintf("  Modified:        %d\n", report.Summary.ModifiedResources))
	sb.WriteString(fmt.Sprintf("  Unchanged:       %d\n\n", report.Summary.UnchangedResources))

	if !HasDrift(report) {
		sb.WriteString("✅ No drift detected!\n")
		return sb.String()
	}

	for _, entry := range report.Entries {
		switch entry.Type {
		case types.DriftAdded:
			sb.WriteString(fmt.Sprintf("  [+] ADDED    %s\n", entry.Resource.FullName()))
		case types.DriftRemoved:
			sb.WriteString(fmt.Sprintf("  [-] REMOVED  %s\n", entry.Resource.FullName()))
		case types.DriftModified:
			sb.WriteString(fmt.Sprintf("  [~] MODIFIED %s\n", entry.Resource.FullName()))
			for _, diff := range entry.FieldDiffs {
				sb.WriteString(fmt.Sprintf("      • %s\n", diff.Path))
				sb.WriteString(fmt.Sprintf("        old: %v\n", diff.OldValue))
				sb.WriteString(fmt.Sprintf("        new: %v\n", diff.NewValue))
			}
		}
	}

	return sb.String()
}

// indexResources creates a map of FullName -> Resource for fast lookup.
func indexResources(resources []types.Resource) map[string]types.Resource {
	index := make(map[string]types.Resource, len(resources))
	for _, r := range resources {
		index[r.FullName()] = r
	}
	return index
}

// compareResources performs a deep comparison of two resources, returning field diffs.
func compareResources(base, target types.Resource) []types.FieldDiff {
	var diffs []types.FieldDiff

	// Compare Labels
	if !reflect.DeepEqual(base.Labels, target.Labels) {
		diffs = append(diffs, types.FieldDiff{
			Path:     ".metadata.labels",
			OldValue: base.Labels,
			NewValue: target.Labels,
		})
	}

	// Compare Annotations
	if !reflect.DeepEqual(base.Annotations, target.Annotations) {
		diffs = append(diffs, types.FieldDiff{
			Path:     ".metadata.annotations",
			OldValue: base.Annotations,
			NewValue: target.Annotations,
		})
	}

	// Compare Spec
	if !reflect.DeepEqual(base.Spec, target.Spec) {
		specDiffs := deepCompareMap(".spec", base.Spec, target.Spec)
		diffs = append(diffs, specDiffs...)
	}

	// Compare Data
	if !reflect.DeepEqual(base.Data, target.Data) {
		dataDiffs := deepCompareMap(".data", base.Data, target.Data)
		diffs = append(diffs, dataDiffs...)
	}

	return diffs
}

// deepCompareMap recursively compares two maps and returns field-level diffs.
func deepCompareMap(prefix string, base, target map[string]interface{}) []types.FieldDiff {
	var diffs []types.FieldDiff

	if base == nil && target == nil {
		return nil
	}

	allKeys := make(map[string]bool)
	for k := range base {
		allKeys[k] = true
	}
	for k := range target {
		allKeys[k] = true
	}

	for k := range allKeys {
		path := prefix + "." + k
		baseVal, baseOk := base[k]
		targetVal, targetOk := target[k]

		if !baseOk {
			diffs = append(diffs, types.FieldDiff{Path: path, NewValue: targetVal})
			continue
		}
		if !targetOk {
			diffs = append(diffs, types.FieldDiff{Path: path, OldValue: baseVal})
			continue
		}

		// Recurse into nested maps
		baseMap, baseIsMap := baseVal.(map[string]interface{})
		targetMap, targetIsMap := targetVal.(map[string]interface{})
		if baseIsMap && targetIsMap {
			diffs = append(diffs, deepCompareMap(path, baseMap, targetMap)...)
			continue
		}

		if !reflect.DeepEqual(baseVal, targetVal) {
			diffs = append(diffs, types.FieldDiff{
				Path:     path,
				OldValue: baseVal,
				NewValue: targetVal,
			})
		}
	}

	return diffs
}
