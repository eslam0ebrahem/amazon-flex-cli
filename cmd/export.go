package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"flexcli/pkg/config"
	"flexcli/pkg/ui"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export authentication tokens to files and print snippets",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintTitle("Token Export")

		atna := config.GetBearerToken()
		atnr := config.GetRefreshToken()

		tokens := config.LoadTokens()
		raw, _ := tokens["raw_response"].(map[string]interface{})
		rawBytes, _ := json.MarshalIndent(raw, "", "  ")
		rawStr := string(rawBytes)
		if len(rawStr) > 3000 {
			rawStr = rawStr[:3000] + "\n… (truncated)"
		}

		// cURL snippet
		curlSnippet := fmt.Sprintf("curl \\\n  -H 'x-amz-access-token: %s' \\\n  -H 'Accept: application/json' \\\n  https://flex-capacity-eu.amazon.com/", atna)
		ui.RenderBox("cURL Snippet", ui.ValueStyle.Render(curlSnippet))

		// Python snippet
		pyToken := atna
		if len(pyToken) > 40 {
			pyToken = pyToken[:40] + "…"
		}
		pySnippet := fmt.Sprintf("headers = {\n    'x-amz-access-token': '%s'\n}", pyToken)
		ui.RenderBox("Python Snippet", ui.ValueStyle.Render(pySnippet))

		// Full JSON preview
		ui.RenderBox("Token JSON (first 3000 chars)", ui.ValueStyle.Render(rawStr))

		// Save to disk
		config.EnsureDbDir()
		if atna != "" {
			p := filepath.Join(config.DbDir, "flex_atna.txt")
			if err := os.WriteFile(p, []byte(atna), 0600); err == nil {
				ui.PrintSuccess("Access token saved to: " + p)
			}
		}
		if atnr != "" {
			p := filepath.Join(config.DbDir, "flex_atnr.txt")
			if err := os.WriteFile(p, []byte(atnr), 0600); err == nil {
				ui.PrintSuccess("Refresh token saved to: " + p)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
