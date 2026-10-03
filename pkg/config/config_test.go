package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestConfig(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test-config.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

func TestLoadConfig(t *testing.T) {
	content := `
cache:
  dir: "./test-cache"
  max_retries: 5
  timeout_seconds: 60

sync:
  interval_minutes: 30
  auto_start: true

rules:
  - name: "test-rule"
    url: "https://example.com/test.srs"
    local_file: "test.srs"
    enabled: true
`
	path := createTestConfig(t, content)

	cfg, err := LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, "./test-cache", cfg.Cache.Dir)
	assert.Equal(t, 5, cfg.Cache.MaxRetries)
	assert.Equal(t, 60, cfg.Cache.TimeoutSeconds)
	assert.Equal(t, 30, cfg.Sync.IntervalMinutes)
	assert.True(t, cfg.Sync.AutoStart)
	assert.Len(t, cfg.Rules, 1)
	assert.Equal(t, "test-rule", cfg.Rules[0].Name)
}

func TestEnvOverrides(t *testing.T) {
	os.Setenv("CACHE_DIR", "/env-cache")
	os.Setenv("MAX_RETRIES", "10")
	os.Setenv("DOWNLOAD_TIMEOUT", "120")
	os.Setenv("SYNC_INTERVAL", "120")
	os.Setenv("AUTO_START", "true")
	defer func() {
		os.Unsetenv("CACHE_DIR")
		os.Unsetenv("MAX_RETRIES")
		os.Unsetenv("DOWNLOAD_TIMEOUT")
		os.Unsetenv("SYNC_INTERVAL")
		os.Unsetenv("AUTO_START")
	}()

	content := `
cache:
  dir: "./default-cache"
rules:
  - name: "test"
    url: "https://example.com/test.srs"
    local_file: "test.srs"
    enabled: true
`
	path := createTestConfig(t, content)

	cfg, err := LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, "/env-cache", cfg.Cache.Dir)
	assert.Equal(t, 10, cfg.Cache.MaxRetries)
	assert.Equal(t, 120, cfg.Cache.TimeoutSeconds)
	assert.Equal(t, 120, cfg.Sync.IntervalMinutes)
	assert.True(t, cfg.Sync.AutoStart)
}

func TestValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			Cache: CacheConfig{
				Dir:            "./cache",
				MaxRetries:     3,
				TimeoutSeconds: 30,
			},
			Rules: []Rule{
				{
					Name:      "test",
					URL:       "https://example.com/test.srs",
					LocalFile: "test.srs",
					Enabled:   true,
				},
			},
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("no rules", func(t *testing.T) {
		cfg := &Config{
			Cache: CacheConfig{TimeoutSeconds: 30},
			Rules: []Rule{},
		}
		assert.Error(t, cfg.Validate())
	})

	t.Run("missing url", func(t *testing.T) {
		cfg := &Config{
			Cache: CacheConfig{TimeoutSeconds: 30},
			Rules: []Rule{
				{
					Name:      "test",
					LocalFile: "test.srs",
					Enabled:   true,
				},
			},
		}
		assert.Error(t, cfg.Validate())
	})

	t.Run("invalid timeout", func(t *testing.T) {
		cfg := &Config{
			Cache: CacheConfig{TimeoutSeconds: -1},
			Rules: []Rule{
				{
					Name:      "test",
					URL:       "https://example.com/test.srs",
					LocalFile: "test.srs",
					Enabled:   true,
				},
			},
		}
		assert.Error(t, cfg.Validate())
	})
}
