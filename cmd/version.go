package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Display the version number of the system monitor application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("System Monitor v%s\n", Version)
	},
}
