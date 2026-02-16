// Package versioner manages Git operations for the snapshot repository.
package versioner

import (
	"fmt"
	"os"
	"time"

	"github.com/raghu-007/GitOps-Time-Machine/pkg/config"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/types"
	log "github.com/sirupsen/logrus"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Versioner manages Git versioning of infrastructure snapshots.
type Versioner struct {
	repoPath string
	config   *config.GitConfig
	repo     *git.Repository
}

// New creates a new Versioner for the given repository path.
func New(repoPath string, cfg *config.GitConfig) (*Versioner, error) {
	v := &Versioner{
		repoPath: repoPath,
		config:   cfg,
	}

	if err := v.initRepo(); err != nil {
		return nil, err
	}

	return v, nil
}

// initRepo initializes or opens the Git repository.
func (v *Versioner) initRepo() error {
	if _, err := os.Stat(v.repoPath); os.IsNotExist(err) {
		if err := os.MkdirAll(v.repoPath, 0755); err != nil {
			return fmt.Errorf("failed to create repo directory: %w", err)
		}
	}

	repo, err := git.PlainOpen(v.repoPath)
	if err != nil {
		// Initialize a new repository
		repo, err = git.PlainInit(v.repoPath, false)
		if err != nil {
			return fmt.Errorf("failed to initialize git repo: %w", err)
		}
		log.WithField("path", v.repoPath).Info("initialized new git repository")

		// Create the configured branch
		if v.config.Branch != "master" {
			headRef := plumbing.NewSymbolicReference(
				plumbing.HEAD,
				plumbing.NewBranchReferenceName(v.config.Branch),
			)
			if err := repo.Storer.SetReference(headRef); err != nil {
				log.WithError(err).Warn("failed to set default branch name")
			}
		}
	}

	v.repo = repo
	return nil
}

// Commit stages all changes and creates a commit with snapshot metadata.
func (v *Versioner) Commit(metadata *types.SnapshotMetadata) (string, error) {
	w, err := v.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Stage all changes
	if err := w.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return "", fmt.Errorf("failed to stage changes: %w", err)
	}

	// Check if there are changes to commit
	status, err := w.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	if status.IsClean() {
		log.Info("no changes detected, skipping commit")
		return "", nil
	}

	// Build commit message
	message := fmt.Sprintf("%s %s — %d resources across %d namespaces",
		v.config.CommitMessagePrefix,
		metadata.Timestamp.Format(time.RFC3339),
		metadata.ResourceCount,
		len(metadata.Namespaces),
	)

	// Create commit
	commit, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  v.config.AuthorName,
			Email: v.config.AuthorEmail,
			When:  metadata.Timestamp,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create commit: %w", err)
	}

	commitObj, err := v.repo.CommitObject(commit)
	if err != nil {
		return "", fmt.Errorf("failed to get commit object: %w", err)
	}

	hash := commitObj.Hash.String()
	log.WithFields(log.Fields{
		"commit":    hash[:8],
		"resources": metadata.ResourceCount,
	}).Info("snapshot committed")

	return hash, nil
}

// History returns the commit log as a list of HistoryEntry.
func (v *Versioner) History(limit int) ([]types.HistoryEntry, error) {
	iter, err := v.repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	var entries []types.HistoryEntry
	count := 0

	err = iter.ForEach(func(c *object.Commit) error {
		if limit > 0 && count >= limit {
			return fmt.Errorf("limit reached")
		}

		entries = append(entries, types.HistoryEntry{
			CommitHash: c.Hash.String(),
			Timestamp:  c.Author.When,
			Message:    c.Message,
			Author:     c.Author.Name,
		})
		count++
		return nil
	})

	// "limit reached" is not a real error
	if err != nil && err.Error() != "limit reached" {
		return nil, err
	}

	return entries, nil
}

// CheckoutAt checks out the snapshot repo at a given commit hash.
func (v *Versioner) CheckoutAt(commitHash string) error {
	w, err := v.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	return w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commitHash),
	})
}

// CheckoutBranch returns to the configured branch.
func (v *Versioner) CheckoutBranch() error {
	w, err := v.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	return w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(v.config.Branch),
	})
}

// FindCommitByTime returns the commit hash closest to (but not after) the given time.
func (v *Versioner) FindCommitByTime(target time.Time) (string, error) {
	iter, err := v.repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get log: %w", err)
	}

	var bestHash string
	var bestTime time.Time

	err = iter.ForEach(func(c *object.Commit) error {
		commitTime := c.Author.When
		if commitTime.Before(target) || commitTime.Equal(target) {
			if bestHash == "" || commitTime.After(bestTime) {
				bestHash = c.Hash.String()
				bestTime = commitTime
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if bestHash == "" {
		return "", fmt.Errorf("no snapshot found at or before %s", target.Format(time.RFC3339))
	}

	return bestHash, nil
}

// GetCommitCount returns the total number of commits in the repository.
func (v *Versioner) GetCommitCount() (int, error) {
	iter, err := v.repo.Log(&git.LogOptions{})
	if err != nil {
		// New repo with no commits
		if err == plumbing.ErrReferenceNotFound {
			return 0, nil
		}
		return 0, err
	}

	count := 0
	_ = iter.ForEach(func(c *object.Commit) error {
		count++
		return nil
	})

	return count, nil
}

// EnsureGitIgnore creates a .gitignore if needed (not required for snapshot repo).
func (v *Versioner) EnsureGitIgnore() error {
	_ = gitconfig.NewConfig() // verify import usage
	return nil
}
