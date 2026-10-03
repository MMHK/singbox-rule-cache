package cli

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/singbox-rule-cache/singbox-rule-cache/pkg/cache"
)

var syncCmd = &cobra.Command{
	Use:   "sync [rule-name]",
	Short: "Manually trigger synchronization",
	Long:  "Download and cache remote rule-set files. Optionally sync a specific rule by name.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSync,
}

func runSync(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	manager := cache.NewManager(cfg)
	ctx := context.Background()

	if len(args) == 1 {
		ruleName := args[0]
		slog.Info("syncing single rule", "name", ruleName)
		return manager.SyncRule(ctx, ruleName)
	}

	slog.Info("syncing all rules")
	return manager.SyncAll(ctx)
}
