package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 定義完整的應用配置結構
type Config struct {
	Cache CacheConfig `yaml:"cache"`
	Sync  SyncConfig  `yaml:"sync"`
	Rules []Rule      `yaml:"rules"`
}

// CacheConfig 定義緩存相關配置
type CacheConfig struct {
	Dir            string `yaml:"dir" env:"CACHE_DIR"`
	MaxRetries     int    `yaml:"max_retries" env:"MAX_RETRIES"`
	TimeoutSeconds int    `yaml:"timeout_seconds" env:"DOWNLOAD_TIMEOUT"`
}

// SyncConfig 定義同步相關配置
type SyncConfig struct {
	IntervalMinutes int  `yaml:"interval_minutes" env:"SYNC_INTERVAL"`
	AutoStart       bool `yaml:"auto_start" env:"AUTO_START"`
}

// LoadConfig 從 YAML 文件加載配置並應用環境變量覆蓋
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	applyDefaults(&cfg)
	applyEnvOverrides(&cfg)

	slog.Info("config loaded", "path", path)
	return &cfg, nil
}

// applyDefaults 應用默認值到未設置的配置項
func applyDefaults(cfg *Config) {
	if cfg.Cache.Dir == "" {
		cfg.Cache.Dir = "./cache"
	}
	if cfg.Cache.MaxRetries == 0 {
		cfg.Cache.MaxRetries = 3
	}
	if cfg.Cache.TimeoutSeconds == 0 {
		cfg.Cache.TimeoutSeconds = 30
	}
	if cfg.Sync.IntervalMinutes == 0 {
		cfg.Sync.IntervalMinutes = 60
	}
}

// applyEnvOverrides 使用環境變量覆蓋配置值
func applyEnvOverrides(cfg *Config) {
	if val := os.Getenv("CACHE_DIR"); val != "" {
		cfg.Cache.Dir = val
	}
	if val := os.Getenv("MAX_RETRIES"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.Cache.MaxRetries = n
		}
	}
	if val := os.Getenv("DOWNLOAD_TIMEOUT"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.Cache.TimeoutSeconds = n
		}
	}
	if val := os.Getenv("SYNC_INTERVAL"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.Sync.IntervalMinutes = n
		}
	}
	if val := os.Getenv("AUTO_START"); val != "" {
		cfg.Sync.AutoStart = strings.ToLower(val) == "true" || val == "1"
	}
}

// Validate 驗證配置的有效性
func (cfg *Config) Validate() error {
	if len(cfg.Rules) == 0 {
		return fmt.Errorf("no rules defined")
	}

	hasEnabled := false
	for i, rule := range cfg.Rules {
		if !rule.Enabled {
			continue
		}
		hasEnabled = true

		if rule.Name == "" {
			return fmt.Errorf("rule[%d]: name is required", i)
		}
		if rule.URL == "" {
			return fmt.Errorf("rule[%d]: url is required", i)
		}
		if rule.LocalFile == "" {
			return fmt.Errorf("rule[%d]: local_file is required", i)
		}
	}

	if !hasEnabled {
		return fmt.Errorf("no enabled rules found")
	}

	if cfg.Cache.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout_seconds must be > 0")
	}

	if cfg.Cache.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be >= 0")
	}

	return nil
}
