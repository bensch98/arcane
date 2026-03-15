package cmd

import (
	"fmt"
	"os"

	"github.com/bensch98/arcane/internal/git"
	"github.com/bensch98/arcane/internal/tracker"
	"github.com/bensch98/arcane/internal/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Pull registry + check outdated items",
	Run: func(cmd *cobra.Command, args []string) {
		// Pull latest
		if git.IsRepo(registryDir) {
			fmt.Println(ui.Bold("Pulling latest registry..."))
			if err := git.Pull(registryDir); err != nil {
				fmt.Println(ui.Yellow("Could not pull (not a git remote or offline)."))
			}
		}

		// Check tracking
		trackingPath := tracker.TrackingFileName
		if _, err := os.Stat(trackingPath); os.IsNotExist(err) {
			fmt.Printf("No %s found. Run 'arcane init' first.\n", trackingPath)
			return
		}

		currentSHA := git.RevParseShort(registryDir)

		fmt.Println()
		fmt.Println(ui.Bold("Installed items:"))

		tf, err := tracker.Load(trackingPath)
		if err != nil {
			ui.Die("Cannot read tracking file: %v", err)
		}

		if len(tf.Installed) == 0 {
			fmt.Println("  No items installed.")
			return
		}

		for _, item := range tf.Installed {
			if currentSHA != "" && item.Version != currentSHA {
				fmt.Printf("  %s (installed: %s, latest: %s) %s\n",
					ui.Yellow(item.Name), item.Version, currentSHA, ui.Yellow("← outdated"))
			} else {
				fmt.Printf("  %s (%s)\n", ui.Green(item.Name), item.Version)
			}
		}

		if currentSHA != "" {
			fmt.Println()
			fmt.Printf("Registry at: %s\n", ui.Dim(currentSHA))
			fmt.Printf("Run %s to update an item.\n", ui.Cyan("arcane add <name> --force"))
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
