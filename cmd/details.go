package cmd

import (
	"fmt"
	"flexcli/pkg/itinerary"
	"flexcli/pkg/ui"

	"github.com/spf13/cobra"
)

var detailsCmd = &cobra.Command{
	Use:   "details [scannable_id]",
	Short: "Get detailed information about a package",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		ui.PrintTitle("Package Details")

		details, err := itinerary.GetDetails(id)
		if err != nil {
			ui.PrintError("Failed to get details: " + err.Error())
			return
		}
		if details == nil {
			ui.PrintError("Package not found: " + id)
			return
		}

		// Identification section
		ui.RenderKeyValueBox("Identification", [][2]string{
			{"Scannable ID", details.ScannableId},
			{"Status", ui.StatusBadge(details.Status)},
			{"Activity Type", details.ActivityType},
			{"Tracking ID", details.TrackingId},
			{"Client Order", details.ClientOrderId},
			{"Weight", details.Weight},
			{"Dimensions", details.Dimensions},
		})

		// Location section
		ui.RenderKeyValueBox("Location & Address", [][2]string{
			{"Coordinates", fmt.Sprintf("%.6f, %.6f", details.Latitude, details.Longitude)},
			{"Recipient", details.RecipientName},
			{"Address", details.AddressName},
			{"City / Postal", details.City + "  " + details.Postal},
			{"Instructions", details.DeliveryInstruct},
			{"Maps URL", fmt.Sprintf("https://maps.google.com/?q=%.6f,%.6f", details.Latitude, details.Longitude)},
		})

		// Images section
		if len(details.Images) > 0 {
			imgRows := [][2]string{}
			for i, img := range details.Images {
				imgRows = append(imgRows, [2]string{fmt.Sprintf("Image %d", i+1), img})
			}
			ui.RenderKeyValueBox("Images", imgRows)
		} else {
			ui.PrintWarning("No images found for this package.")
		}
	},
}

func init() {
	rootCmd.AddCommand(detailsCmd)
}
