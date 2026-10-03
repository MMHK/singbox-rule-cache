package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/singbox-rule-cache/singbox-rule-cache/pkg/cache"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List cached rule status",
	Long:  "Display the status of all configured rules and their cache state.",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	manager := cache.NewManager(cfg)
	statuses, err := manager.ListCache()
	if err != nil {
		return fmt.Errorf("list cache: %w", err)
	}

	printStatusTable(statuses)
	return nil
}

func printStatusTable(statuses []cache.CacheStatus) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	fmt.Fprintln(w, "NAME\tURL\tEXISTS\tSIZE\tLAST UPDATED\tSTATUS")
	fmt.Fprintln(w, "----\t---\t------\t----\t------------\t------")

	for _, s := range statuses {
		exists := "No"
		size := "-"
		lastUpdated := "-"
		validationStatus := s.ValidationStatus

		if s.Exists {
			exists = "Yes"
			size = formatSize(s.FileSize)
			lastUpdated = s.LastUpdated.Format(time.RFC3339)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			s.Name,
			truncateURL(s.URL, 40),
			exists,
			size,
			lastUpdated,
			validationStatus,
		)
	}

	w.Flush()
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}
