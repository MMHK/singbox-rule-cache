package cli

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/singbox-rule-cache/singbox-rule-cache/pkg/cache"
	"github.com/singbox-rule-cache/singbox-rule-cache/pkg/config"
)

var intervalMinutes int

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start periodic synchronization",
	Long:  "Start the syncer to periodically download and cache remote rule-set files.",
	RunE:  runStart,
}

func init() {
	startCmd.Flags().IntVar(&intervalMinutes, "interval", 0, "sync interval in minutes (overrides config)")
}

func runStart(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if intervalMinutes > 0 {
		cfg.Sync.IntervalMinutes = intervalMinutes
	}

	manager := cache.NewManager(cfg)
	syncer := cache.NewSyncer(manager, cfg.Sync.IntervalMinutes)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := syncer.Start(ctx); err != nil {
		return err
	}

	slog.Info("press Ctrl+C to stop")
	<-ctx.Done()

	slog.Info("shutting down...")
	return syncer.Stop()
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	if cacheDir != "" {
		cfg.Cache.Dir = cacheDir
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
