package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sysmon/system-monitor-cli/internal/dev"
)

var (
	snapshotFormat string
	snapshotOutput string
)

func init() {
	snapshotCmd.Flags().StringVarP(&snapshotFormat, "format", "f", "markdown", "output format (markdown or json)")
	snapshotCmd.Flags().StringVarP(&snapshotOutput, "output", "o", "", "file path to write snapshot to")
	rootCmd.AddCommand(snapshotCmd)
}

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Generate a developer environment & system diagnostic report",
	Long: `Inspect and report system specifications, developer runtime versions,
active listening ports, and memory consumption. Output can be formatted as Markdown
or JSON to share in bug reports or performance audits.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		snap, err := dev.GenerateSnapshot()
		if err != nil {
			return fmt.Errorf("failed to generate snapshot: %w", err)
		}

		var content string
		if snapshotFormat == "json" {
			data, err := json.MarshalIndent(snap, "", "  ")
			if err != nil {
				return err
			}
			content = string(data)
		} else {
			content = snap.FormatMarkdown()
		}

		if snapshotOutput != "" {
			if err := os.WriteFile(snapshotOutput, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write snapshot to %s: %w", snapshotOutput, err)
			}
			fmt.Printf("Snapshot saved to %s\n", snapshotOutput)
			return nil
		}

		fmt.Println(content)
		return nil
	},
}
