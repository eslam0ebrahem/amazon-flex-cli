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
	simLat, simLon float64
	picLat, picLon float64
)

var simulateCmd = &cobra.Command{
	Use:   "simulate [scannable_id]",
	Short: "Simulate a full pickup workflow without sending the request",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		ui.PrintTitle("Simulate Pickup")
		ui.PrintInfo(fmt.Sprintf("Building payload preview for: %s", id))

		data := itinerary.LoadItinerary()
		if !data.Exists() {
			ui.PrintError("No itinerary loaded — run flexcli packages to fetch first.")
			return
		}

		act, op, tr, item, found := itinerary.FindPackage(data, id)
		if !found {
			ui.PrintError("Package not found: " + id)
			return
		}

		actType := act.Get("activityType").String()
		trID := tr.Get("id").String()
		ui.PrintSuccess(fmt.Sprintf("Found — Activity: %s  |  TR: %s", actType, trID))

		lat, lon := simLat, simLon
		if lat == 0 && lon == 0 {
			var ok bool
			lat, lon, ok = itinerary.GetPackageLocation(act, op)
			if !ok {
				ui.PrintWarning("No location found — using default (30.13384, 31.72604)")
				lat, lon = 30.13384, 31.72604
			} else {
				ui.PrintSuccess(fmt.Sprintf("Auto-detected location: %.6f, %.6f", lat, lon))
			}
		} else {
			ui.PrintInfo(fmt.Sprintf("Using custom location: %.6f, %.6f", lat, lon))
		}

		images := itinerary.GetPackageImages(tr, item)
		ui.PrintInfo(fmt.Sprintf("Images attached: %d", len(images)))

		newLat, newLon := utils.GenerateGPSInCircle(lat, lon, 3.0)
		ui.PrintSuccess(fmt.Sprintf("Spoofed GPS (3m radius): %.6f, %.6f", newLat, newLon))

		nowEpoch := float64(time.Now().UnixNano()) / 1e9
		payload := api.BuildPickupPayload(id, trID, item.Get("id").String(), newLat, newLon, nowEpoch)

		previewBytes, _ := json.MarshalIndent(payload, "", "  ")
		preview := string(previewBytes)
		if len(preview) > 900 {
			preview = preview[:900] + "\n… (truncated)"
		}

		ui.RenderBox("Payload Preview", ui.ValueStyle.Render(preview))
		ui.PrintInfo("SIMULATION only — no request was sent.")
		ui.PrintInfo(fmt.Sprintf("To execute: flexcli pickup %s", id))
	},
}

var pickupCmd = &cobra.Command{
	Use:   "pickup [scannable_id]",
	Short: "Pick up a package automatically via API",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		ui.PrintTitle("Execute Pickup")
		ui.PrintInfo(fmt.Sprintf("Preparing pickup for: %s", id))

		data := itinerary.LoadItinerary()
		if !data.Exists() {
			ui.PrintError("No itinerary loaded — run flexcli packages to fetch first.")
			return
		}

		act, op, tr, item, found := itinerary.FindPackage(data, id)
		if !found {
			ui.PrintError("Package not found: " + id)
			return
		}

		lat, lon := picLat, picLon
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
		payload := api.BuildPickupPayload(id, trID, item.Get("id").String(), newLat, newLon, nowEpoch)

		ui.PrintSuccess(fmt.Sprintf("Spoofed GPS: %.6f, %.6f  |  TR: %s", newLat, newLon, trID))
		ui.PrintInfo("Sending pickup request…")

		resp, err := api.RecordPickup(payload)
		if err != nil {
			ui.PrintError("Pickup API call failed: " + err.Error())
			return
		}

		respBytes, _ := json.MarshalIndent(resp, "", "  ")
		ui.RenderBox("API Response", ui.ValueStyle.Render(string(respBytes)))
		ui.PrintSuccess("Pickup requested successfully!")
	},
}

func init() {
	simulateCmd.Flags().Float64Var(&simLat, "lat", 0, "Custom latitude to spoof from")
	simulateCmd.Flags().Float64Var(&simLon, "lon", 0, "Custom longitude to spoof from")

	pickupCmd.Flags().Float64Var(&picLat, "lat", 0, "Custom latitude to spoof from")
	pickupCmd.Flags().Float64Var(&picLon, "lon", 0, "Custom longitude to spoof from")

	rootCmd.AddCommand(simulateCmd)
	rootCmd.AddCommand(pickupCmd)
}
