package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bensch98/arcane/internal/installer"
	"github.com/bensch98/arcane/internal/tracker"
	"github.com/bensch98/arcane/internal/ui"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove installed item",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		trackingPath := tracker.TrackingFileName
		if _, err := os.Stat(trackingPath); os.IsNotExist(err) {
			ui.Die("No %s found. Run 'arcane init' first or remove files manually.", trackingPath)
		}

		tf, err := tracker.Load(trackingPath)
		if err != nil {
			ui.Die("Cannot read tracking file: %v", err)
		}

		tracked := tracker.FindInstalled(tf, name)
		if tracked == nil {
			ui.Die("Item '%s' is not tracked in %s.", name, trackingPath)
		}

		fmt.Printf("%s %s", ui.Bold("Removing:"), name)
		if tracked.Tool != "" {
			fmt.Printf(" [%s]", tracked.Tool)
		}
		fmt.Println()

		// Delete tracked files
		for _, f := range tracked.Files {
			if _, err := os.Stat(f); err == nil {
				os.Remove(f)
				fmt.Printf("  %s %s\n", ui.Red("deleted:"), f)
			} else {
				fmt.Printf("  %s %s\n", ui.Dim("not found:"), f)
			}
		}

		// Handle hook removal (Claude Code)
		item := reg.FindItem(name)
		if item != nil && item.Type == "hook" && item.HookMerge != nil {
			tool := tracked.Tool
			if tool == "" {
				tool = "claude" // backward compat
			}
			settingsFile := reg.SettingsFileForTool(tool, "hook")
			if settingsFile == "" {
				settingsFile = ".claude/settings.json"
			}
			settingsPath := filepath.Join(".", settingsFile)
			if _, err := os.Stat(settingsPath); err == nil {
				if err := installer.RemoveHook(settingsPath, item.HookMerge); err != nil {
					fmt.Printf("  %s could not clean hook: %v\n", ui.Yellow("warning:"), err)
				} else {
					fmt.Printf("  %s %s from %s\n", ui.Red("removed hook:"), item.HookMerge.Event, settingsPath)
				}
			}
		}

		// Handle config removal (OpenCode)
		if item != nil && (item.Type == "formatter" || item.Type == "config-merge") && item.ConfigMerge != nil {
			tool := tracked.Tool
			if tool == "" {
				tool = "opencode"
			}
			configFile := reg.SettingsFileForTool(tool, item.Type)
			if configFile == "" {
				configFile = "opencode.json"
			}
			configPath := filepath.Join(".", configFile)
			if _, err := os.Stat(configPath); err == nil {
				if err := installer.RemoveConfig(configPath, item.ConfigMerge); err != nil {
					fmt.Printf("  %s could not clean config: %v\n", ui.Yellow("warning:"), err)
				} else {
					fmt.Printf("  %s %s from %s\n", ui.Red("removed config:"), item.ConfigMerge.Path, configPath)
				}
			}
		}

		// Remove from tracking
		if err := tracker.Untrack(trackingPath, name); err != nil {
			ui.Die("Cannot update tracking file: %v", err)
		}

		fmt.Println(ui.Green("Done."))
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
