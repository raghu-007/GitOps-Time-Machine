// Package scheduler provides cron-based scheduling for periodic snapshots.
package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

// SnapshotFunc is the function that will be called on each scheduled tick.
type SnapshotFunc func(ctx context.Context) error

// Scheduler manages periodic snapshot execution.
type Scheduler struct {
	cron       *cron.Cron
	schedule   string
	snapshotFn SnapshotFunc
	mu         sync.Mutex
	running    bool
	cancelFn   context.CancelFunc
}

// New creates a new Scheduler with the given cron schedule.
func New(schedule string, fn SnapshotFunc) (*Scheduler, error) {
	// Validate the cron expression
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	if _, err := parser.Parse(schedule); err != nil {
		return nil, fmt.Errorf("invalid cron schedule %q: %w", schedule, err)
	}

	return &Scheduler{
		cron:       cron.New(),
		schedule:   schedule,
		snapshotFn: fn,
	}, nil
}

// Start begins the scheduled snapshot execution.
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.mu.Unlock()

	childCtx, cancel := context.WithCancel(ctx)
	s.cancelFn = cancel

	_, err := s.cron.AddFunc(s.schedule, func() {
		log.Info("scheduler: triggering snapshot")
		if err := s.snapshotFn(childCtx); err != nil {
			log.WithError(err).Error("scheduler: snapshot failed")
		} else {
			log.Info("scheduler: snapshot completed successfully")
		}
	})
	if err != nil {
		cancel()
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.cron.Start()
	log.WithField("schedule", s.schedule).Info("scheduler started")

	// Block until context is cancelled
	<-childCtx.Done()

	log.Info("scheduler: stopping...")
	cronCtx := s.cron.Stop()
	<-cronCtx.Done()

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	log.Info("scheduler stopped")
	return nil
}

// Stop halts the scheduler.
func (s *Scheduler) Stop() {
	if s.cancelFn != nil {
		s.cancelFn()
	}
}

// IsRunning returns whether the scheduler is currently active.
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
