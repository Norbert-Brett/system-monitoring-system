package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/sysmon/system-monitor-cli/internal/dev"
)

var (
	portsKillTarget int
	portsForceKill  bool
	portsJSONOutput bool
	portsDevOnly    bool
)

func init() {
	portsCmd.Flags().IntVarP(&portsKillTarget, "kill", "k", 0, "kill process listening on this port")
	portsCmd.Flags().BoolVarP(&portsForceKill, "force", "f", false, "force kill (SIGKILL)")
	portsCmd.Flags().BoolVar(&portsJSONOutput, "json", false, "output ports as JSON")
	portsCmd.Flags().BoolVar(&portsDevOnly, "dev-only", false, "only display developer-related services")

	rootCmd.AddCommand(portsCmd)
}

var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Inspect listening ports and processes (with quick-kill support)",
	Long: `List all listening network ports along with their PID, process name,
developer tech stack, and project directory.

Use --kill <port> to quickly terminate a process holding a port.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If --kill flag is provided, kill the specified port
		if portsKillTarget > 0 {
			killed, err := dev.KillPort(portsKillTarget, portsForceKill)
			if err != nil {
				return fmt.Errorf("failed to kill port %d: %w", portsKillTarget, err)
			}
			successColor := color.New(color.FgGreen, color.Bold)
			fmt.Printf("%s Port :%d (PID %d: %s) has been terminated.\n",
				successColor.Sprint("✔ SUCCESS:"), killed.Port, killed.PID, killed.ProcessName)
			return nil
		}

		ports, err := dev.GetListeningPorts()
		if err != nil {
			return fmt.Errorf("failed to retrieve listening ports: %w", err)
		}

		if portsDevOnly {
			var filtered []dev.ListeningPort
			for _, p := range ports {
				if p.Category != "System" && p.Category != "" {
					filtered = append(filtered, p)
				}
			}
			ports = filtered
		}

		if portsJSONOutput {
			data, err := json.MarshalIndent(ports, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		if len(ports) == 0 {
			fmt.Println("No active listening ports found.")
			return nil
		}

		headerColor := color.New(color.FgCyan, color.Bold)
		portColor := color.New(color.FgYellow, color.Bold)
		stackColor := color.New(color.FgGreen)
		dimColor := color.New(color.FgHiBlack)

		fmt.Println()
		headerColor.Printf("  %-8s %-8s %-20s %-16s %-15s %s\n",
			"PORT", "PID", "PROCESS", "STACK", "CATEGORY", "PROJECT / CWD")
		dimColor.Println("  " + "------------------------------------------------------------------------------------------------")

		for _, p := range ports {
			loc := p.Project
			if loc == "" {
				loc = p.CWD
			}
			if loc == "" {
				loc = "-"
			}

			stack := p.Stack
			if stack == "" {
				stack = "System"
			}
			category := p.Category
			if category == "" {
				category = "System"
			}

			fmt.Printf("  :%-7s %-8d %-20s %-16s %-15s %s\n",
				portColor.Sprintf("%d", p.Port),
				p.PID,
				p.ProcessName,
				stackColor.Sprint(stack),
				category,
				loc,
			)
		}
		fmt.Printf("\n  Total: %d listening ports. (Run 'sysmon kill-port <port>' to free a port)\n\n", len(ports))
		return nil
	},
}
