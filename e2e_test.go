package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/singbox-rule-cache/singbox-rule-cache/pkg/cache"
	"github.com/singbox-rule-cache/singbox-rule-cache/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_DownloadAndValidateAllRules 測試下載並驗證所有規則
func TestE2E_DownloadAndValidateAllRules(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	tmpDir := t.TempDir()
	cfg := createTestConfig(tmpDir)

	manager := cache.NewManager(cfg)
	err := manager.SyncAll(ctx)

	require.NoError(t, err, "SyncAll should succeed")

	for _, rule := range cfg.Rules {
		localPath := filepath.Join(tmpDir, rule.LocalFile)
		assert.FileExists(t, localPath, "SRS file should exist: %s", rule.Name)

		metaPath := localPath + ".meta.json"
		assert.FileExists(t, metaPath, "Metadata file should exist: %s", rule.Name)

		err := cache.ValidateSRSFile(localPath)
		assert.NoError(t, err, "SRS validation should pass: %s", rule.Name)
	}
}

// TestE2E_ConditionalRequest 測試條件請求（304 Not Modified）
func TestE2E_ConditionalRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	tmpDir := t.TempDir()
	cfg := createTestConfig(tmpDir)
	manager := cache.NewManager(cfg)

	err := manager.SyncAll(ctx)
	require.NoError(t, err, "First sync should succeed")

	firstSizes := getFileSizes(t, tmpDir, cfg)

	time.Sleep(2 * time.Second)

	err = manager.SyncAll(ctx)
	require.NoError(t, err, "Second sync should succeed")

	secondSizes := getFileSizes(t, tmpDir, cfg)
	assert.Equal(t, firstSizes, secondSizes, "File sizes should remain unchanged after conditional request")
}

// TestE2E_ListCacheStatus 測試列出緩存狀態
func TestE2E_ListCacheStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	tmpDir := t.TempDir()
	cfg := createTestConfig(tmpDir)
	manager := cache.NewManager(cfg)

	err := manager.SyncAll(ctx)
	require.NoError(t, err, "Sync should succeed")

	statuses, err := manager.ListCache()
	require.NoError(t, err, "ListCache should succeed")
	assert.Len(t, statuses, len(cfg.Rules), "Should return status for all rules")

	for _, status := range statuses {
		assert.True(t, status.Exists, "Rule should exist: %s", status.Name)
		assert.Greater(t, status.FileSize, int64(0), "File size should be > 0: %s", status.Name)
		assert.Equal(t, "valid", status.ValidationStatus, "Validation should be valid: %s", status.Name)
		assert.False(t, status.LastUpdated.IsZero(), "LastUpdated should be set: %s", status.Name)
	}
}

// TestE2E_CleanInvalidCache 測試清理無效緩存
func TestE2E_CleanInvalidCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	tmpDir := t.TempDir()
	cfg := createTestConfig(tmpDir)
	manager := cache.NewManager(cfg)

	err := manager.SyncAll(ctx)
	require.NoError(t, err, "Initial sync should succeed")

	firstRule := cfg.Rules[0]
	corruptPath := filepath.Join(tmpDir, firstRule.LocalFile)
	err = os.WriteFile(corruptPath, []byte("invalid data"), 0644)
	require.NoError(t, err, "Should be able to corrupt file")

	err = manager.CleanInvalid()
	require.NoError(t, err, "CleanInvalid should succeed")

	assert.NoFileExists(t, corruptPath, "Corrupted file should be removed")
	assert.NoFileExists(t, corruptPath+".meta.json", "Metadata should also be removed")
}

// createTestConfig 創建測試專用配置
func createTestConfig(cacheDir string) *config.Config {
	return &config.Config{
		Cache: config.CacheConfig{
			Dir:            cacheDir,
			MaxRetries:     3,
			TimeoutSeconds: 30,
		},
		Sync: config.SyncConfig{
			IntervalMinutes: 60,
			AutoStart:       false,
		},
		Rules: []config.Rule{
			{
				Name:      "overseas-ai",
				URL:       "https://raw.githubusercontent.com/mm-sam/OverseasAI.list/refs/heads/main/rule/Singbox/OverseasAI/OverseasAI.srs",
				LocalFile: "OverseasAI.srs",
				Enabled:   true,
			},
			{
				Name:      "ip-private",
				URL:       "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geoip/private.srs",
				LocalFile: "ip-private.srs",
				Enabled:   true,
			},
			{
				Name:      "site-private",
				URL:       "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geosite/private.srs",
				LocalFile: "site-private.srs",
				Enabled:   true,
			},
			{
				Name:      "category-ads-all",
				URL:       "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geosite/category-ads-all.srs",
				LocalFile: "category-ads-all.srs",
				Enabled:   true,
			},
			{
				Name:      "accelerated-domains-china",
				URL:       "https://raw.githubusercontent.com/Dreista/sing-box-rule-set-cn/rule-set/accelerated-domains.china.conf.srs",
				LocalFile: "accelerated-domains.china.conf.srs",
				Enabled:   true,
			},
			{
				Name:      "ip-cn",
				URL:       "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geoip/cn.srs",
				LocalFile: "ip-cn.srs",
				Enabled:   true,
			},
			{
				Name:      "site-cn",
				URL:       "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geosite/cn.srs",
				LocalFile: "site-cn.srs",
				Enabled:   true,
			},
			{
				Name:      "geolocation-not-cn",
				URL:       "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geosite/geolocation-!cn.srs",
				LocalFile: "geolocation-!cn.srs",
				Enabled:   true,
			},
		},
	}
}

// getFileSizes 獲取所有規則文件的大小映射
func getFileSizes(t *testing.T, cacheDir string, cfg *config.Config) map[string]int64 {
	sizes := make(map[string]int64)
	for _, rule := range cfg.Rules {
		localPath := filepath.Join(cacheDir, rule.LocalFile)
		info, err := os.Stat(localPath)
		require.NoError(t, err, "Should be able to stat file: %s", rule.Name)
		sizes[rule.Name] = info.Size()
	}
	return sizes
}
