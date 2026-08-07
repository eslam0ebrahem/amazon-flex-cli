package cmd

import (
	"fmt"
	"flexcli/pkg/itinerary"
	"flexcli/pkg/ui"

	"github.com/spf13/cobra"
)

var locateCmd = &cobra.Command{
	Use:   "locate [scannable_id]",
	Short: "Get GPS coordinates and map link for a package",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		ui.PrintTitle("Package Location")

		details, err := itinerary.GetDetails(id)
		if err != nil {
			ui.PrintError("Failed to get details: " + err.Error())
			return
		}
		if details == nil {
			ui.PrintError("Package not found: " + id)
			return
		}

		if details.Latitude == 0 && details.Longitude == 0 {
			ui.PrintError("No location data found for package " + id)
			return
		}

		mapsURL := fmt.Sprintf("https://maps.google.com/?q=%.6f,%.6f", details.Latitude, details.Longitude)
		ui.RenderKeyValueBox("GPS Coordinates", [][2]string{
			{"Scannable ID", details.ScannableId},
			{"Latitude", fmt.Sprintf("%.6f", details.Latitude)},
			{"Longitude", fmt.Sprintf("%.6f", details.Longitude)},
			{"Recipient", details.RecipientName},
			{"Address", details.AddressName + ", " + details.City},
			{"Google Maps", mapsURL},
		})
		ui.PrintSuccess("Coordinates found.")
	},
}

func init() {
	rootCmd.AddCommand(locateCmd)
}
