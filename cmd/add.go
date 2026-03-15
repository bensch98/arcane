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

var validTypes = map[string]bool{
	"command": true,
	"script":  true,
	"skill":   true,
	"hook":    true,
}

var addCmd = &cobra.Command{
	Use:   "add <type> <name...> | add all | add sync",
	Short: "Install items + dependencies",
	Long: `Install registry items by type and name.

Examples:
  arcane add command commit-message i18n
  arcane add hook stop-notify-toast post-edit-prettier
  arcane add script notify-toast-script
  arcane add all                          # install every item
  arcane add sync                         # reinstall items from .arcane.json`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		first := args[0]

		var names []string

		switch first {
		case "all":
			for _, item := range reg.Items {
				names = append(names, item.Name)
			}
		case "sync":
			names = syncFromTracking()
		default:
			if !validTypes[first] {
				ui.Die("Unknown type '%s'. Use: command, script, skill, hook, all, or sync.", first)
			}
			if len(args) < 2 {
				ui.Die("Usage: arcane add %s <name...>", first)
			}
			typ := first
			for _, name := range args[1:] {
				item := reg.FindItem(name)
				if item == nil {
					ui.Die("Item '%s' not found in registry.", name)
				}
				if item.Type != typ {
					ui.Die("Item '%s' is a %s, not a %s.", name, item.Type, typ)
				}
				names = append(names, name)
			}
		}

		if len(names) == 0 {
			fmt.Println("Nothing to install.")
			return
		}

		installItems(names)
	},
}

func syncFromTracking() []string {
	trackingPath := tracker.TrackingFileName
	if _, err := os.Stat(trackingPath); os.IsNotExist(err) {
		ui.Die("No %s found. Run 'arcane init' first.", trackingPath)
	}
	tf, err := tracker.Load(trackingPath)
	if err != nil {
		ui.Die("Cannot read tracking file: %v", err)
	}
	if len(tf.Installed) == 0 {
		ui.Die("No items in %s.", trackingPath)
	}
	var names []string
	for _, item := range tf.Installed {
		if reg.FindItem(item.Name) != nil {
			names = append(names, item.Name)
		} else {
			fmt.Printf("  %s '%s' not found in registry, skipping\n", ui.Yellow("warning:"), item.Name)
		}
	}
	return names
}

func installItems(names []string) {
	items, err := reg.ResolveMultipleDeps(names)
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
	// Track files per top-level requested name
	filesByName := make(map[string][]string)

	for _, itemName := range items {
		item := reg.FindItem(itemName)
		fmt.Printf("  %s (%s)\n", ui.Cyan(itemName), item.Type)

		files, err := installer.InstallItem(item, registryDir, targetRoot, addForce, addDryRun, rb)
		if err != nil {
			rb.Undo()
			ui.Die("Installation failed for '%s': %v", itemName, err)
		}
		filesByName[itemName] = files
	}

	// Update tracking
	if !addDryRun && !addGlobal {
		if _, err := os.Stat(tracker.TrackingFileName); err == nil {
			sha := git.RevParseShort(registryDir)
			if sha == "" {
				sha = "unknown"
			}
			trackingPath := filepath.Join(targetRoot, tracker.TrackingFileName)
			for _, name := range names {
				// Collect all files for this name and its deps
				resolved, _ := reg.ResolveDeps(name)
				var allFiles []string
				for _, dep := range resolved {
					allFiles = append(allFiles, filesByName[dep]...)
				}
				if err := tracker.Track(trackingPath, name, sha, allFiles); err != nil {
					fmt.Printf("  %s could not update tracking for '%s': %v\n", ui.Yellow("warning:"), name, err)
				}
			}
		}
	}

	fmt.Println()
	fmt.Println(ui.Green("Done."))
}

func init() {
	addCmd.Flags().BoolVar(&addGlobal, "global", false, "Install to $HOME instead of current directory")
	addCmd.Flags().BoolVar(&addForce, "force", false, "Overwrite existing files")
	addCmd.Flags().BoolVar(&addDryRun, "dry-run", false, "Show what would be installed without writing")
	rootCmd.AddCommand(addCmd)
}
