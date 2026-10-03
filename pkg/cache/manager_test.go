package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/singbox-rule-cache/singbox-rule-cache/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestManager(t *testing.T, rules []config.Rule) *Manager {
	t.Helper()

	tmpDir := t.TempDir()
	cfg := &config.Config{
		Cache: config.CacheConfig{
			Dir:            tmpDir,
			MaxRetries:     2,
			TimeoutSeconds: 10,
		},
		Rules: rules,
	}

	return NewManager(cfg)
}

func TestSyncRule(t *testing.T) {
	rules := []config.Rule{
		{
			Name:      "test-rule",
			URL:       "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geoip/private.srs",
			LocalFile: "private.srs",
			Enabled:   true,
		},
	}

	mgr := createTestManager(t, rules)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := mgr.SyncRule(ctx, "test-rule")
	assert.NoError(t, err)

	// Verify file exists
	localPath := filepath.Join(mgr.config.Cache.Dir, "private.srs")
	_, err = os.Stat(localPath)
	assert.NoError(t, err)
}

func TestSyncRuleNotFound(t *testing.T) {
	rules := []config.Rule{
		{
			Name:      "existing-rule",
			URL:       "https://example.com/test.srs",
			LocalFile: "test.srs",
			Enabled:   true,
		},
	}

	mgr := createTestManager(t, rules)

	ctx := context.Background()
	err := mgr.SyncRule(ctx, "nonexistent-rule")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule not found")
}

func TestListCache(t *testing.T) {
	rules := []config.Rule{
		{
			Name:      "rule-1",
			URL:       "https://example.com/rule1.srs",
			LocalFile: "rule1.srs",
			Enabled:   true,
		},
		{
			Name:      "rule-2",
			URL:       "https://example.com/rule2.srs",
			LocalFile: "rule2.srs",
			Enabled:   false,
		},
	}

	mgr := createTestManager(t, rules)

	statuses, err := mgr.ListCache()
	require.NoError(t, err)
	assert.Len(t, statuses, 2)

	// Check first rule status (should not exist yet)
	assert.Equal(t, "rule-1", statuses[0].Name)
	assert.False(t, statuses[0].Exists)
	assert.Equal(t, "", statuses[0].ValidationStatus)
}

func TestSyncAllWithInvalidURL(t *testing.T) {
	rules := []config.Rule{
		{
			Name:      "invalid-rule",
			URL:       "https://invalid.example.com/nonexistent.srs",
			LocalFile: "invalid.srs",
			Enabled:   true,
		},
	}

	mgr := createTestManager(t, rules)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Should fail since all rules are invalid
	err := mgr.SyncAll(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all rules failed")
}

func TestCleanInvalid(t *testing.T) {
	rules := []config.Rule{
		{
			Name:      "test-rule",
			URL:       "https://example.com/test.srs",
			LocalFile: "test.srs",
			Enabled:   true,
		},
	}

	mgr := createTestManager(t, rules)

	// Create an invalid file
	localPath := filepath.Join(mgr.config.Cache.Dir, "test.srs")
	err := os.WriteFile(localPath, []byte("invalid content"), 0644)
	require.NoError(t, err)

	// Clean should remove the invalid file
	err = mgr.CleanInvalid()
	assert.NoError(t, err)

	// Verify file was removed
	_, err = os.Stat(localPath)
	assert.True(t, os.IsNotExist(err))
}
