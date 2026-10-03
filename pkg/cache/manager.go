package cache

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/singbox-rule-cache/singbox-rule-cache/pkg/config"
)

// Manager coordinates concurrent downloads of multiple SRS files
type Manager struct {
	config     *config.Config
	downloader *Downloader
}

// NewManager creates a new Manager instance
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config:     cfg,
		downloader: NewDownloader(cfg.Cache.MaxRetries, cfg.Cache.TimeoutSeconds),
	}
}

// SyncAll downloads all enabled rules concurrently
func (m *Manager) SyncAll(ctx context.Context) error {
	slog.Info("starting sync of all rules")

	successCount, failCount := m.syncRulesConcurrently(ctx)

	slog.Info("sync completed", "success", successCount, "failed", failCount)

	if failCount > 0 && successCount == 0 {
		return fmt.Errorf("all rules failed to sync")
	}

	if failCount > 0 {
		slog.Warn("some rules failed during sync", "failed_count", failCount)
	}

	return nil
}

// syncRulesConcurrently launches goroutines for all enabled rules
func (m *Manager) syncRulesConcurrently(ctx context.Context) (int, int) {
	var wg sync.WaitGroup
	errChan := make(chan error, len(m.config.Rules))
	var successCount, failCount atomic.Int64

	for _, rule := range m.config.Rules {
		if !rule.Enabled {
			continue
		}

		wg.Add(1)
		go func(r config.Rule) {
			defer wg.Done()
			if err := m.syncSingleRule(ctx, r); err != nil {
				slog.Warn("failed to sync rule", "name", r.Name, "error", err)
				errChan <- fmt.Errorf("rule %s: %w", r.Name, err)
				failCount.Add(1)
			} else {
				slog.Debug("rule synced successfully", "name", r.Name)
				successCount.Add(1)
			}
		}(rule)
	}

	wg.Wait()
	close(errChan)

	return int(successCount.Load()), int(failCount.Load())
}

// syncSingleRule downloads a single rule with proper path handling
func (m *Manager) syncSingleRule(ctx context.Context, rule config.Rule) error {
	localPath := filepath.Join(m.config.Cache.Dir, rule.LocalFile)

	slog.Debug("downloading rule", "name", rule.Name, "url", rule.URL, "path", localPath)

	meta, err := m.downloader.Download(ctx, rule.URL, localPath)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	slog.Info("rule downloaded", "name", rule.Name, "size", meta.FileSize)
	return nil
}

// SyncRule downloads a single rule by name
func (m *Manager) SyncRule(ctx context.Context, ruleName string) error {
	rule, err := m.findRuleByName(ruleName)
	if err != nil {
		return err
	}

	slog.Info("syncing single rule", "name", ruleName)
	return m.syncSingleRule(ctx, rule)
}

// CacheStatus represents the status of a cached rule
type CacheStatus struct {
	Name             string    `json:"name"`
	URL              string    `json:"url"`
	LocalFile        string    `json:"local_file"`
	Exists           bool      `json:"exists"`
	LastUpdated      time.Time `json:"last_updated"`
	FileSize         int64     `json:"file_size"`
	ValidationStatus string    `json:"validation_status"`
}

// ListCache returns status of all cached rules
func (m *Manager) ListCache() ([]CacheStatus, error) {
	statuses := make([]CacheStatus, 0, len(m.config.Rules))

	for _, rule := range m.config.Rules {
		status := m.getRuleStatus(rule)
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// getRuleStatus retrieves status for a single rule
func (m *Manager) getRuleStatus(rule config.Rule) CacheStatus {
	localPath := filepath.Join(m.config.Cache.Dir, rule.LocalFile)
	status := CacheStatus{
		Name:      rule.Name,
		URL:       rule.URL,
		LocalFile: localPath,
	}

	meta, err := loadMetadata(localPath)
	if err != nil {
		slog.Debug("no metadata found", "rule", rule.Name, "error", err)
		return status
	}

	fileInfo, err := os.Stat(localPath)
	if err == nil {
		status.Exists = true
		status.FileSize = fileInfo.Size()
		status.LastUpdated = meta.DownloadedAt
		status.ValidationStatus = meta.ValidationStatus
	}

	return status
}

// CleanInvalid removes cache files that failed validation
func (m *Manager) CleanInvalid() error {
	slog.Info("cleaning invalid cache files")

	removedCount := 0
	for _, rule := range m.config.Rules {
		localPath := filepath.Join(m.config.Cache.Dir, rule.LocalFile)
		if err := m.cleanInvalidRule(localPath); err != nil {
			slog.Warn("failed to clean rule", "name", rule.Name, "error", err)
		} else {
			removedCount++
		}
	}

	slog.Info("cleanup completed", "removed", removedCount)
	return nil
}

// cleanInvalidRule checks and removes invalid cache for a single rule
func (m *Manager) cleanInvalidRule(localPath string) error {
	if err := ValidateSRSFile(localPath); err != nil {
		slog.Warn("removing invalid file", "path", localPath, "error", err)
		os.Remove(localPath)
		os.Remove(localPath + ".meta.json")
		return nil
	}
	return nil
}

// findRuleByName finds a rule by its name
func (m *Manager) findRuleByName(name string) (config.Rule, error) {
	for _, rule := range m.config.Rules {
		if rule.Name == name {
			return rule, nil
		}
	}
	return config.Rule{}, fmt.Errorf("rule not found: %s", name)
}
