// Package timetravel provides time-travel query functionality for infrastructure snapshots.
package timetravel

import (
	"fmt"
	"time"

	"github.com/raghu-007/GitOps-Time-Machine/pkg/snapshotter"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/types"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/versioner"
	log "github.com/sirupsen/logrus"
)

// Engine provides time-travel queries over the snapshot Git repository.
type Engine struct {
	versioner   *versioner.Versioner
	snapshotter *snapshotter.Snapshotter
	repoPath    string
}

// New creates a new time-travel Engine.
func New(v *versioner.Versioner, s *snapshotter.Snapshotter, repoPath string) *Engine {
	return &Engine{
		versioner:   v,
		snapshotter: s,
		repoPath:    repoPath,
	}
}

// SnapshotAt retrieves the infrastructure state at a given time.
func (e *Engine) SnapshotAt(target time.Time) (*types.ResourceSnapshot, error) {
	log.WithField("target", target.Format(time.RFC3339)).Info("time-travel: looking up snapshot")

	// Find the commit closest to the target time
	commitHash, err := e.versioner.FindCommitByTime(target)
	if err != nil {
		return nil, fmt.Errorf("failed to find snapshot at %s: %w", target.Format(time.RFC3339), err)
	}

	return e.SnapshotByCommit(commitHash)
}

// SnapshotByCommit retrieves the infrastructure state at a specific commit.
func (e *Engine) SnapshotByCommit(commitHash string) (*types.ResourceSnapshot, error) {
	log.WithField("commit", commitHash[:8]).Info("time-travel: checking out snapshot")

	// Checkout the commit
	if err := e.versioner.CheckoutAt(commitHash); err != nil {
		return nil, fmt.Errorf("failed to checkout commit %s: %w", commitHash, err)
	}

	// Ensure we return to the branch when done
	defer func() {
		if err := e.versioner.CheckoutBranch(); err != nil {
			log.WithError(err).Warn("failed to return to branch")
		}
	}()

	// Read the snapshot at this commit
	snapshot, err := e.snapshotter.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot at commit %s: %w", commitHash, err)
	}

	snapshot.Metadata.CommitHash = commitHash
	return snapshot, nil
}

// CompareTimeRange compares infrastructure state between two points in time.
func (e *Engine) CompareTimeRange(from, to time.Time) (*types.ResourceSnapshot, *types.ResourceSnapshot, error) {
	log.WithFields(log.Fields{
		"from": from.Format(time.RFC3339),
		"to":   to.Format(time.RFC3339),
	}).Info("time-travel: comparing time range")

	fromSnapshot, err := e.SnapshotAt(from)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get snapshot at 'from' time: %w", err)
	}

	toSnapshot, err := e.SnapshotAt(to)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get snapshot at 'to' time: %w", err)
	}

	return fromSnapshot, toSnapshot, nil
}

// ListResources returns all resources at a given time matching optional filters.
func (e *Engine) ListResources(target time.Time, kind string, namespace string) ([]types.Resource, error) {
	snapshot, err := e.SnapshotAt(target)
	if err != nil {
		return nil, err
	}

	var filtered []types.Resource
	for _, res := range snapshot.Resources {
		if kind != "" && res.Kind != kind {
			continue
		}
		if namespace != "" && res.Namespace != namespace {
			continue
		}
		filtered = append(filtered, res)
	}

	return filtered, nil
}
