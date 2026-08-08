package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"flexcli/pkg/api"

	"github.com/spf13/cobra"
)

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Manage device settings",
	Long:  "Commands to manage the active device for the Amazon Flex account.",
}

var activeDeviceCmd = &cobra.Command{
	Use:   "active",
	Short: "Get the currently active device",
	Run: func(cmd *cobra.Command, args []string) {
		device, err := api.GetActiveDevice()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(device, "", "  ")
		fmt.Println("Active Device:")
		fmt.Println(string(data))
	},
}

var setActiveDeviceCmd = &cobra.Command{
	Use:   "set-active",
	Short: "Set the CLI as the active device",
	Run: func(cmd *cobra.Command, args []string) {
		err := api.SetActiveDevice()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Successfully set CLI as the active device!")
	},
}

var deviceNonceCmd = &cobra.Command{
	Use:   "nonce",
	Short: "Request an App Attestation Nonce from Amazon",
	Run: func(cmd *cobra.Command, args []string) {
		res, err := api.GetAppAttestationNonce()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println("App Attestation Nonce Response:")
		fmt.Println(string(data))
	},
}

func init() {
	rootCmd.AddCommand(deviceCmd)
	deviceCmd.AddCommand(activeDeviceCmd)
	deviceCmd.AddCommand(setActiveDeviceCmd)
	deviceCmd.AddCommand(deviceNonceCmd)
}

