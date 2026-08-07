package cmd

import (
	"encoding/json"
	"flexcli/pkg/api"
	"flexcli/pkg/ui"

	"github.com/spf13/cobra"
)

var driverCmd = &cobra.Command{
	Use:   "driver",
	Short: "Manage driver profile and routes",
}

var updatePhoneCmd = &cobra.Command{
	Use:   "phone [phone_number]",
	Short: "Update your active work phone number on file with Amazon Flex",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		phone := args[0]
		ui.PrintTitle("Update Work Phone")
		ui.PrintInfo("Sending phone number update to Amazon Flex...")

		err := api.UpdateWorkPhone(phone)
		if err != nil {
			ui.PrintError("Failed to update phone: " + err.Error())
			return
		}

		ui.PrintSuccess("Phone number successfully updated to: " + phone)
	},
}

var assignmentsCmd = &cobra.Command{
	Use:   "assignments",
	Short: "List scheduled assignments (blocks)",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintTitle("Scheduled Assignments")
		ui.PrintInfo("Fetching scheduled blocks...")

		data, err := api.GetScheduledAssignments()
		if err != nil {
			ui.PrintError("Failed to fetch assignments: " + err.Error())
			return
		}

		b, _ := json.MarshalIndent(data, "", "  ")
		ui.PrintSuccess("Assignments retrieved:")
		println(string(b))
	},
}

var toursCmd = &cobra.Command{
	Use:   "tours",
	Short: "List associated route tour IDs",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintTitle("Active Tours")
		ui.PrintInfo("Fetching active tours...")

		data, err := api.GetAssociatedTRs()
		if err != nil {
			ui.PrintError("Failed to fetch tours: " + err.Error())
			return
		}

		b, _ := json.MarshalIndent(data, "", "  ")
		ui.PrintSuccess("Tours retrieved:")
		println(string(b))
	},
}

var instantOffersCmd = &cobra.Command{
	Use:   "instant-offers",
	Short: "Check Real-Time Availability (Instant Offers) status",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintTitle("Instant Offers Status")
		ui.PrintInfo("Checking availability...")

		data, err := api.GetRealTimeAvailability()
		if err != nil {
			ui.PrintError("Failed to check status: " + err.Error())
			return
		}

		b, _ := json.MarshalIndent(data, "", "  ")
		ui.PrintSuccess("Status retrieved:")
		println(string(b))
	},
}

func init() {
	driverCmd.AddCommand(updatePhoneCmd)
	driverCmd.AddCommand(assignmentsCmd)
	driverCmd.AddCommand(toursCmd)
	driverCmd.AddCommand(instantOffersCmd)
	rootCmd.AddCommand(driverCmd)
}

