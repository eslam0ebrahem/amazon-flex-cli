package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"flexcli/pkg/api"
	"flexcli/pkg/itinerary"
	"flexcli/pkg/ui"
	"flexcli/pkg/utils"

	"github.com/spf13/cobra"
)

var (
	updLat, updLon float64
)

var statusMap = map[string]struct {
	Reason string
	State  string
}{
	"pickup":              {"NONE", "PICKED_UP"},
	"closed":              {"BUSINESS_CLOSED", "PICKUP_FAILED"},
	"no-address":          {"ADDRESS_NOT_FOUND", "PICKUP_FAILED"},
	"customer-reschedule": {"RESCHEDULED_BY_CUSTOMER", "PICKUP_FAILED"},
	"out-of-time":         {"OUT_OF_DELIVERY_TIME", "PICKUP_FAILED"},
}

var updatePackageCmd = &cobra.Command{
	Use:   "update [scannable_id] [status_key]",
	Short: "Update package state using predefined keys (e.g. pickup, closed)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		key := args[1]

		status, exists := statusMap[key]
		if !exists {
			var keys []string
			for k := range statusMap {
				keys = append(keys, k)
			}
			ui.PrintError(fmt.Sprintf("Invalid status key '%s'. Available keys: %v", key, keys))
			return
		}

		reason := status.Reason
		state := status.State

		ui.PrintTitle("Update Package")
		ui.PrintInfo(fmt.Sprintf("Preparing update for: %s -> %s (%s)", id, state, reason))

		data := itinerary.LoadItinerary()
		if !data.Exists() {
			ui.PrintError("No itinerary loaded — run flexcli packages to fetch first.")
			return
		}

		act, op, tr, item, found := itinerary.FindPackage(data, id)
		if !found {
			ui.PrintError("Package not found in itinerary: " + id)
			return
		}

		lat, lon := updLat, updLon
		if lat == 0 && lon == 0 {
			var ok bool
			lat, lon, ok = itinerary.GetPackageLocation(act, op)
			if !ok {
				lat, lon = 30.13384, 31.72604
			}
		}

		newLat, newLon := utils.GenerateGPSInCircle(lat, lon, 3.0)
		nowEpoch := float64(time.Now().UnixNano()) / 1e9
		trID := tr.Get("id").String()
		itemId := item.Get("id").String()

		payload := api.BuildUpdatePayload(id, trID, itemId, newLat, newLon, nowEpoch, reason, state)

		ui.PrintSuccess(fmt.Sprintf("Auto-resolved TR: %s | Item: %s", trID, itemId))
		ui.PrintSuccess(fmt.Sprintf("Spoofed GPS: %.6f, %.6f", newLat, newLon))
		ui.PrintInfo("Sending update request...")

		resp, err := api.RecordPickup(payload)
		if err != nil {
			ui.PrintError("Update API call failed: " + err.Error())
			return
		}

		respBytes, _ := json.MarshalIndent(resp, "", "  ")
		ui.RenderBox("API Response", ui.ValueStyle.Render(string(respBytes)))
		ui.PrintSuccess("Package updated successfully!")
	},
}

func init() {
	updatePackageCmd.Flags().Float64Var(&updLat, "lat", 0, "Custom latitude to spoof from")
	updatePackageCmd.Flags().Float64Var(&updLon, "lon", 0, "Custom longitude to spoof from")

	rootCmd.AddCommand(updatePackageCmd)
}
