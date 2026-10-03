package cli

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/singbox-rule-cache/singbox-rule-cache/pkg/cache"
)

var validateCmd = &cobra.Command{
	Use:   "validate [file-path]",
	Short: "Validate cached SRS files",
	Long:  "Validate the integrity of cached SRS files. Optionally validate a specific file.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		filePath := args[0]
		return validateSingleFile(filePath)
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	manager := cache.NewManager(cfg)
	statuses, err := manager.ListCache()
	if err != nil {
		return fmt.Errorf("list cache: %w", err)
	}

	validCount := 0
	invalidCount := 0

	for _, status := range statuses {
		if !status.Exists {
			slog.Info("file not found", "name", status.Name)
			continue
		}

		if err := cache.ValidateSRSFile(status.LocalFile); err != nil {
			slog.Warn("validation failed", "name", status.Name, "error", err)
			invalidCount++
		} else {
			slog.Info("validation passed", "name", status.Name)
			validCount++
		}
	}

	slog.Info("validation completed", "valid", validCount, "invalid", invalidCount)
	return nil
}

func validateSingleFile(filePath string) error {
	slog.Info("validating file", "path", filePath)

	if err := cache.ValidateSRSFile(filePath); err != nil {
		slog.Error("validation failed", "error", err)
		return fmt.Errorf("validation failed: %w", err)
	}

	slog.Info("validation passed")
	return nil
}
