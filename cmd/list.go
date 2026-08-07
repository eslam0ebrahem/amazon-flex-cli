package cmd

import (
	"fmt"
	"flexcli/pkg/itinerary"
	"flexcli/pkg/ui"

	"github.com/charmbracelet/bubbles/table"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all packages in the itinerary",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintTitle("Itinerary — Package List")

		data := itinerary.LoadItinerary()
		if !data.Exists() {
			ui.PrintError("No itinerary loaded — run flexcli packages to fetch first.")
			return
		}

		var rows []table.Row
		activities := itinerary.ExtractActivities(data)
		for _, act := range activities {
			actType := act.Get("activityType").String()
			for _, op := range act.Get("operations").Array() {
				for _, tr := range op.Get("transportRequests").Array() {
					trScannable := tr.Get("scannableId").String()
					trackingId := tr.Get("clientMetaData.trackingId").String()
					extObjId := tr.Get("clientMetaData.externalObjectId").String()
					slamTracking := tr.Get("labels.SLAM.details.trackingId.text").String()
					slamAltExec := tr.Get("labels.SLAM.details.alternateExecutionId.text").String()
					trScannables := []string{trScannable, trackingId, extObjId, slamTracking, slamAltExec}

					for _, item := range tr.Get("transportItems").Array() {
						sid := ""
						for _, s := range append([]string{item.Get("scannableId").String()}, trScannables...) {
							if s != "" {
								sid = s
								break
							}
						}

						state := item.Get("state").String()
						if state == "" {
							state = tr.Get("transportObjectState").String()
						}

						if sid != "" {
							rows = append(rows, table.Row{
								sid,
								state,
								actType,
								tr.Get("id").String(),
							})
						}
					}
				}
			}
		}

		if len(rows) == 0 {
			ui.PrintError("No packages found in itinerary.")
			return
		}

		columns := []table.Column{
			{Title: "Scannable ID", Width: 22},
			{Title: "Status", Width: 20},
			{Title: "Activity", Width: 15},
			{Title: "TR ID", Width: 45},
		}

		ui.RenderBox("Packages", ui.BuildTable(columns, rows))
		ui.PrintSuccess(fmt.Sprintf("Found %d packages.", len(rows)))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
