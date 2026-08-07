package cmd

import (
	"flexcli/pkg/api"
	"flexcli/pkg/config"
	"flexcli/pkg/itinerary"
	"flexcli/pkg/ui"

	"github.com/spf13/cobra"
)

var packagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "Fetch and update the itinerary from Amazon Flex API",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintTitle("Fetch Itinerary")

		token := config.GetBearerToken()
		if token == "" {
			ui.PrintError("No auth token found — run flexcli login first.")
			return
		}

		ui.PrintInfo("Connecting to Amazon Flex API…")

		data, err := api.FetchItinerary(0, 0)
		if err != nil {
			ui.PrintError("API error: " + err.Error())
			return
		}
		if err = itinerary.SaveItinerary(data); err != nil {
			ui.PrintError("Failed to save itinerary: " + err.Error())
			return
		}

		config.EnsureDbDir()
		ui.PrintSuccess("Itinerary fetched and saved to: " + config.ItineraryFile)
		ui.PrintInfo("Run flexcli list to see your packages, or flexcli dashboard for the full TUI.")
	},
}

func init() {
	rootCmd.AddCommand(packagesCmd)
}
