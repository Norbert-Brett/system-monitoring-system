package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/sysmon/system-monitor-cli/internal/dev"
)

var runJSONOutput bool

func init() {
	runCmd.Flags().BoolVar(&runJSONOutput, "json", false, "output profile results as JSON")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run <command> [args...]",
	Short: "Execute and profile a build, test, or dev command",
	Long: `Run any shell command or test suite while tracking peak RSS memory usage,
CPU execution time, and total duration.

Examples:
  sysmon run go test ./...
  sysmon run -- npm run build
  sysmon run --json cargo test`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var cleanArgs []string
		jsonOutput := false

		for _, a := range args {
			if a == "--json" {
				jsonOutput = true
			} else if a == "--help" || a == "-h" {
				return cmd.Help()
			} else {
				cleanArgs = append(cleanArgs, a)
			}
		}

		if len(cleanArgs) == 0 {
			return cmd.Help()
		}

		command := cleanArgs[0]
		cmdArgs := cleanArgs[1:]

		bannerColor := color.New(color.FgCyan, color.Bold)
		if !jsonOutput {
			bannerColor.Printf("\n🚀 Profiling command: %s %s\n\n", command, strings.Join(cmdArgs, " "))
		}

		profile, err := dev.ProfileCommand(context.Background(), command, cmdArgs, nil)
		if err != nil {
			return err
		}

		if jsonOutput {
			data, err := json.MarshalIndent(profile, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		titleColor := color.New(color.FgYellow, color.Bold)
		metricColor := color.New(color.FgWhite, color.Bold)
		statusColor := color.New(color.FgGreen, color.Bold)
		if profile.ExitCode != 0 {
			statusColor = color.New(color.FgRed, color.Bold)
		}

		fmt.Println("\n" + strings.Repeat("─", 60))
		titleColor.Println("📊 Execution & Resource Summary:")
		fmt.Printf("  Status:       %s\n", statusColor.Sprintf("Exit Code %d", profile.ExitCode))
		fmt.Printf("  Duration:     %s\n", metricColor.Sprint(profile.Duration.Round(10*time.Millisecond)))
		fmt.Printf("  Peak RSS:     %s (%.2f MB)\n",
			metricColor.Sprintf("%d bytes", profile.PeakRSS),
			float64(profile.PeakRSS)/(1024*1024),
		)
		fmt.Printf("  User CPU:     %s\n", metricColor.Sprint(profile.UserCPU.Round(time.Millisecond)))
		fmt.Printf("  System CPU:   %s\n", metricColor.Sprint(profile.SysCPU.Round(time.Millisecond)))
		fmt.Println(strings.Repeat("─", 60) + "\n")

		if profile.ExitCode != 0 {
			os.Exit(profile.ExitCode)
		}
		return nil
	},
}
