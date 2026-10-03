package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Syncer manages periodic synchronization of SRS files
type Syncer struct {
	manager  *Manager
	interval time.Duration
	ticker   *time.Ticker
	stopChan chan struct{}
	running  bool
	mu       sync.Mutex
}

// NewSyncer creates a new Syncer instance
func NewSyncer(manager *Manager, intervalMinutes int) *Syncer {
	if intervalMinutes <= 0 {
		intervalMinutes = 30 // default to 30 minutes
	}

	return &Syncer{
		manager:  manager,
		interval: time.Duration(intervalMinutes) * time.Minute,
		stopChan: make(chan struct{}),
		running:  false,
	}
}

// Start begins the periodic synchronization loop
func (s *Syncer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("syncer is already running")
	}

	s.ticker = time.NewTicker(s.interval)
	s.stopChan = make(chan struct{})
	s.running = true
	s.mu.Unlock()

	slog.Info("syncer started", "interval", s.interval)

	go s.run(ctx)

	return nil
}

// Stop stops the periodic synchronization
func (s *Syncer) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil // already stopped, no error
	}

	s.running = false
	s.mu.Unlock()

	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)

	slog.Info("syncer stopped")
	return nil
}

// SyncNow triggers an immediate synchronization
func (s *Syncer) SyncNow(ctx context.Context) error {
	slog.Info("manual sync triggered")

	if err := s.manager.SyncAll(ctx); err != nil {
		slog.Warn("manual sync failed", "error", err)
		return fmt.Errorf("sync failed: %w", err)
	}

	slog.Info("manual sync completed successfully")
	return nil
}

// IsRunning returns whether the syncer is currently running
func (s *Syncer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// run executes the main synchronization loop
func (s *Syncer) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("syncer stopped via context")
			return
		case <-s.stopChan:
			slog.Info("syncer stopped via stop channel")
			return
		case <-s.ticker.C:
			slog.Debug("ticker triggered, starting sync")
			if err := s.manager.SyncAll(ctx); err != nil {
				slog.Warn("periodic sync failed", "error", err)
			} else {
				slog.Info("periodic sync completed successfully")
			}
		}
	}
}
