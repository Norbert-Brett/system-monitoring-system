package cmd

import (
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/sysmon/system-monitor-cli/internal/dev"
)

var forceKillPort bool

func init() {
	killPortCmd.Flags().BoolVarP(&forceKillPort, "force", "f", false, "force kill (SIGKILL)")
	rootCmd.AddCommand(killPortCmd)
}

var killPortCmd = &cobra.Command{
	Use:   "kill-port <port>",
	Short: "Quickly kill the process occupying a specific port",
	Long: `Terminate the process holding a network port, freeing it for development.
Sends SIGTERM by default and falls back to SIGKILL if necessary.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := strconv.Atoi(args[0])
		if err != nil || port <= 0 || port > 65535 {
			return fmt.Errorf("invalid port number %q (must be 1-65535)", args[0])
		}

		killed, err := dev.KillPort(port, forceKillPort)
		if err != nil {
			return err
		}

		successColor := color.New(color.FgGreen, color.Bold)
		fmt.Printf("\n%s Successfully freed port :%d\n", successColor.Sprint("✔"), killed.Port)
		fmt.Printf("  Process: %s\n", killed.ProcessName)
		fmt.Printf("  PID:     %d\n", killed.PID)
		if killed.Project != "" {
			fmt.Printf("  Project: %s\n", killed.Project)
		} else if killed.CWD != "" {
			fmt.Printf("  CWD:     %s\n", killed.CWD)
		}
		fmt.Println()
		return nil
	},
}
