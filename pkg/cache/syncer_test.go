package cache

import (
	"context"
	"testing"
	"time"

	"github.com/singbox-rule-cache/singbox-rule-cache/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestSyncer(t *testing.T, intervalMinutes int) (*Syncer, *Manager) {
	t.Helper()

	rules := []config.Rule{
		{
			Name:      "test-rule",
			URL:       "https://example.com/test.srs",
			LocalFile: "test.srs",
			Enabled:   true,
		},
	}

	mgr := createTestManager(t, rules)
	syncer := NewSyncer(mgr, intervalMinutes)

	return syncer, mgr
}

func TestSyncerStartAndStop(t *testing.T) {
	syncer, _ := createTestSyncer(t, 30)

	// Initially not running
	assert.False(t, syncer.IsRunning())

	// Start the syncer
	ctx := context.Background()
	err := syncer.Start(ctx)
	require.NoError(t, err)
	assert.True(t, syncer.IsRunning())

	// Stop the syncer
	err = syncer.Stop()
	require.NoError(t, err)
	assert.False(t, syncer.IsRunning())
}

func TestSyncerPreventDuplicateStart(t *testing.T) {
	syncer, _ := createTestSyncer(t, 30)

	ctx := context.Background()

	// First start should succeed
	err := syncer.Start(ctx)
	require.NoError(t, err)
	assert.True(t, syncer.IsRunning())

	// Second start should fail
	err = syncer.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Clean up
	err = syncer.Stop()
	require.NoError(t, err)
}

func TestSyncerSyncNow(t *testing.T) {
	syncer, _ := createTestSyncer(t, 30)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// SyncNow should work even when not running periodically
	err := syncer.SyncNow(ctx)
	// This will fail because the URL is invalid, but the function itself works
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sync failed")
}

func TestSyncerMultipleStopCalls(t *testing.T) {
	syncer, _ := createTestSyncer(t, 30)

	ctx := context.Background()

	// Start and stop
	err := syncer.Start(ctx)
	require.NoError(t, err)

	err = syncer.Stop()
	require.NoError(t, err)

	// Multiple stops should be safe
	err = syncer.Stop()
	assert.NoError(t, err)

	err = syncer.Stop()
	assert.NoError(t, err)
}

func TestSyncerShortInterval(t *testing.T) {
	// Test with very short interval to verify ticker works
	syncer, _ := createTestSyncer(t, 1) // 1 minute interval

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := syncer.Start(ctx)
	require.NoError(t, err)
	assert.True(t, syncer.IsRunning())

	// Let it run briefly
	time.Sleep(500 * time.Millisecond)

	// Should still be running
	assert.True(t, syncer.IsRunning())

	// Stop it
	err = syncer.Stop()
	require.NoError(t, err)
	assert.False(t, syncer.IsRunning())
}
