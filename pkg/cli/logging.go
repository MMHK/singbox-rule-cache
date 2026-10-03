package cli

import (
	"log/slog"
	"os"
)

// setupLogging configures the slog logger based on verbosity flag
func setupLogging() {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}
