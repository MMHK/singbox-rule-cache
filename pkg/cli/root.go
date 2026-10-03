package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	configPath string
	cacheDir   string
	verbose    bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "singbox-rule-cache",
	Short: "Cache sing-box remote rule-set .srs files",
	Long:  "A tool to cache sing-box remote rule-set .srs files to local storage.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogging()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "./config.yaml", "config file path")
	rootCmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", "", "cache directory (overrides config)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose logging")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(listCmd)
}
