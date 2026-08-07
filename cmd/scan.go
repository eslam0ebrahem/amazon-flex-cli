package cmd

import (
	"fmt"
	"flexcli/pkg/itinerary"
	"flexcli/pkg/ui"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [scannable_id]",
	Short: "Deep search across all possible IDs for a package",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		ui.PrintTitle("Deep Scan")
		ui.PrintInfo(fmt.Sprintf("Searching all ID fields for: %s", id))

		data := itinerary.LoadItinerary()
		if !data.Exists() {
			ui.PrintError("No itinerary loaded — run flexcli packages to fetch first.")
			return
		}

		act, _, tr, item, found := itinerary.FindPackage(data, id)
		if !found {
			ui.PrintError("Package not found across any fields for ID: " + id)
			return
		}

		ui.RenderKeyValueBox("Scan Results", [][2]string{
			{"Found in TR", tr.Get("transportRequestId").String()},
			{"Item ID", item.Get("id").String()},
			{"Activity Type", act.Get("activityType").String()},
			{"Scannable ID", item.Get("scannableId").String()},
			{"Tracking ID", tr.Get("clientMetaData.trackingId").String()},
		})
		ui.PrintSuccess("Scan complete — ID resolves to the above package.")
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
