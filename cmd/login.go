package cmd

import (
	"flexcli/pkg/api"
	"flexcli/pkg/config"
	"flexcli/pkg/ui"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [email] [password]",
	Short: "Authenticate with Amazon Flex",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		password := args[1]

		ui.PrintTitle("Login — Amazon Flex")
		ui.PrintInfo("Connecting to Amazon Flex authentication servers…")

		data, err := api.Login(email, password)
		if err != nil {
			ui.PrintError("Login failed: " + err.Error())
			return
		}

		if err = config.SaveTokens(data, email, password); err != nil {
			ui.PrintError("Failed to save tokens: " + err.Error())
			return
		}

		ui.PrintSuccess("Authentication successful — tokens saved.")
	},
}

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the Amazon Flex authentication token",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintTitle("Token Refresh")
		ui.PrintInfo("Refreshing access token…")

		data, err := api.Refresh()
		if err != nil {
			ui.PrintError("Refresh failed: " + err.Error())
			return
		}

		if err = config.SaveTokens(data, "", ""); err != nil {
			ui.PrintError("Failed to save refreshed tokens: " + err.Error())
			return
		}

		ui.PrintSuccess("Token refreshed successfully.")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintTitle("Authentication Status")

		token := config.GetBearerToken()
		if token == "" {
			ui.PrintError("Not logged in — run flexcli login first.")
			return
		}

		preview := token
		if len(token) > 20 {
			preview = token[:20] + "…"
		}

		emailStr := "Loading..."
		if email, err := api.GetUserProfile(); err == nil {
			emailStr = email
		} else {
			emailStr = "Error: " + err.Error()
		}

		ui.RenderKeyValueBox("Session", [][2]string{
			{"State", "✔  Logged in"},
			{"Email", emailStr},
			{"Access Token", preview},
			{"Session ID", config.GetSessionId()},
		})
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(refreshCmd)
	rootCmd.AddCommand(statusCmd)
}
