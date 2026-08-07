package cmd

import (
	"fmt"
	"flexcli/pkg/itinerary"
	"flexcli/pkg/ui"

	"github.com/spf13/cobra"
)

var imagesCmd = &cobra.Command{
	Use:   "images [scannable_id]",
	Short: "List image URLs for a package",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		ui.PrintTitle("Package Images")

		details, err := itinerary.GetDetails(id)
		if err != nil {
			ui.PrintError("Failed to get details: " + err.Error())
			return
		}
		if details == nil {
			ui.PrintError("Package not found: " + id)
			return
		}

		if len(details.Images) == 0 {
			ui.PrintWarning("No images found for package " + id)
			return
		}

		rows := [][2]string{}
		for i, img := range details.Images {
			rows = append(rows, [2]string{fmt.Sprintf("Image %d", i+1), img})
		}
		ui.RenderKeyValueBox(fmt.Sprintf("Images (%d)", len(details.Images)), rows)
		ui.PrintSuccess(fmt.Sprintf("Found %d image(s) for %s.", len(details.Images), id))
	},
}

func init() {
	rootCmd.AddCommand(imagesCmd)
}
