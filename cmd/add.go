package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bensch98/arcane/internal/git"
	"github.com/bensch98/arcane/internal/installer"
	"github.com/bensch98/arcane/internal/tracker"
	"github.com/bensch98/arcane/internal/ui"
	"github.com/spf13/cobra"
)

var addGlobal bool
var addForce bool
var addDryRun bool

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Install item + dependencies",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if reg.FindItem(name) == nil {
			ui.Die("Item '%s' not found in registry.", name)
		}

		items, err := reg.ResolveDeps(name)
		if err != nil {
			ui.Die("%v", err)
		}

		targetRoot := "."
		if addGlobal {
			home, _ := os.UserHomeDir()
			targetRoot = home
		}

		fmt.Printf("%s %s\n", ui.Bold("Installing:"), strings.Join(items, " "))
		if addDryRun {
			fmt.Println(ui.Yellow("(dry run — no files will be written)"))
		}
		fmt.Println()

		rb := installer.NewRollback()
		var allInstalledFiles []string

		for _, itemName := range items {
			item := reg.FindItem(itemName)
			fmt.Printf("  %s (%s)\n", ui.Cyan(itemName), item.Type)

			files, err := installer.InstallItem(item, registryDir, targetRoot, addForce, addDryRun, rb)
			if err != nil {
				rb.Undo()
				ui.Die("Installation failed for '%s': %v", itemName, err)
			}
			allInstalledFiles = append(allInstalledFiles, files...)
		}

		// Update tracking
		if !addDryRun && !addGlobal {
			if _, err := os.Stat(tracker.TrackingFileName); err == nil {
				sha := git.RevParseShort(registryDir)
				if sha == "" {
					sha = "unknown"
				}
				trackingPath := filepath.Join(targetRoot, tracker.TrackingFileName)
				if err := tracker.Track(trackingPath, name, sha, allInstalledFiles); err != nil {
					fmt.Printf("  %s could not update tracking: %v\n", ui.Yellow("warning:"), err)
				}
			}
		}

		fmt.Println()
		fmt.Println(ui.Green("Done."))
	},
}

func init() {
	addCmd.Flags().BoolVar(&addGlobal, "global", false, "Install to $HOME instead of current directory")
	addCmd.Flags().BoolVar(&addForce, "force", false, "Overwrite existing files")
	addCmd.Flags().BoolVar(&addDryRun, "dry-run", false, "Show what would be installed without writing")
	rootCmd.AddCommand(addCmd)
}
