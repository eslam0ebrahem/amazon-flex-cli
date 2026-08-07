package cmd

import (
	"fmt"
	"os"

	"flexcli/pkg/ui"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "flexcli",
	Short: "Amazon Flex Advanced Operations Manager",
	Long: `A modern, high-performance CLI for managing Amazon Flex API interactions.
Use "flexcli dashboard" for the interactive TUI, or individual subcommands.`,
	// Print the banner when running with no arguments.
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintBanner()
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
