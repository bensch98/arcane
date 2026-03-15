package cmd

import (
	"fmt"
	"os"

	"github.com/bensch98/arcane/internal/tracker"
	"github.com/bensch98/arcane/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create .arcane.json tracking file",
	Run: func(cmd *cobra.Command, args []string) {
		path := tracker.TrackingFileName
		if _, err := os.Stat(path); err == nil {
			fmt.Println(ui.Yellow(path + " already exists."))
			return
		}

		if err := os.WriteFile(path, []byte("{\n  \"installed\": []\n}\n"), 0644); err != nil {
			ui.Die("Cannot create %s: %v", path, err)
		}
		fmt.Println(ui.Green("Created " + path))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
