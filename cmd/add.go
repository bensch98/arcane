package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bensch98/arcane/internal/git"
	"github.com/bensch98/arcane/internal/installer"
	"github.com/bensch98/arcane/internal/registry"
	"github.com/bensch98/arcane/internal/tracker"
	"github.com/bensch98/arcane/internal/ui"
	"github.com/spf13/cobra"
)

var addGlobal bool
var addForce bool
var addDryRun bool
var addToolFlag string

var addCmd = &cobra.Command{
	Use:   "add <type> <name...> | add all | add sync",
	Short: "Install items + dependencies",
	Long: `Install registry items by type and name.

Examples:
  arcane add command commit-message i18n
  arcane add hook stop-notify-toast post-edit-prettier
  arcane add script notify-toast-script
  arcane add all                          # install every item for detected tool
  arcane add sync                         # reinstall items from .arcane.json
  arcane add --tool=opencode command commit-message`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		first := args[0]

		// Resolve target tools
		targetTools := resolveTargetTools()
		if len(targetTools) == 0 {
			ui.Die("No target tool detected. Create a .claude/ or .opencode/ directory, or use --tool.")
		}

		var names []string

		switch first {
		case "all":
			// Collect all item names that are compatible with at least one target tool
			seen := map[string]bool{}
			for _, tool := range targetTools {
				for _, item := range reg.ItemsForTool(tool) {
					if !seen[item.Name] {
						seen[item.Name] = true
						names = append(names, item.Name)
					}
				}
			}
		case "sync":
			names = syncFromTracking()
		default:
			// Validate type against all target tools
			validTypes := mergedValidTypes(targetTools)
			if !validTypes[first] {
				typeList := sortedKeys(validTypes)
				ui.Die("Unknown type '%s'. Use: %s, all, or sync.", first, strings.Join(typeList, ", "))
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
				// Verify the item is compatible with at least one target tool
				compatible := false
				for _, tool := range targetTools {
					if item.Tool.Contains(tool) {
						compatible = true
						break
					}
				}
				if !compatible {
					ui.Die("Item '%s' is for %s, not %s.", name, item.Tool.String(), strings.Join(targetTools, "/"))
				}
				names = append(names, name)
			}
		}

		if len(names) == 0 {
			fmt.Println("Nothing to install.")
			return
		}

		installItems(names, targetTools)
	},
}

// resolveTargetTools determines which tool(s) to install for.
func resolveTargetTools() []string {
	// Explicit --tool flag takes priority
	if addToolFlag != "" {
		// Validate that the tool exists in the registry
		if _, ok := reg.Tools[addToolFlag]; !ok {
			toolList := sortedKeys(reg.Tools)
			ui.Die("Unknown tool '%s'. Available: %s", addToolFlag, strings.Join(toolList, ", "))
		}
		return []string{addToolFlag}
	}

	// Auto-detect from project directories
	targetRoot := "."
	if addGlobal {
		home, _ := os.UserHomeDir()
		targetRoot = home
	}

	detected := registry.DetectTools(targetRoot)

	if len(detected) == 0 {
		// No tool directories found — default to claude for backward compat
		return []string{"claude"}
	}
	if len(detected) == 1 {
		return detected
	}

	// Multiple tools detected — ask user
	sort.Strings(detected)
	fmt.Printf("%s Multiple tool directories found: %s\n", ui.Bold("?"), strings.Join(detected, ", "))
	fmt.Printf("  Install for which tool? [%s/all] ", strings.Join(detected, "/"))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "all" || input == "" {
		return detected
	}

	// Validate the user's choice
	for _, tool := range detected {
		if tool == input {
			return []string{input}
		}
	}

	ui.Die("Invalid choice '%s'. Expected one of: %s, all", input, strings.Join(detected, ", "))
	return nil
}

// mergedValidTypes returns valid types across all target tools.
func mergedValidTypes(tools []string) map[string]bool {
	merged := make(map[string]bool)
	for _, tool := range tools {
		for k, v := range reg.ValidTypesForTool(tool) {
			if v {
				merged[k] = true
			}
		}
	}
	return merged
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

func installItems(names []string, targetTools []string) {
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
	fmt.Printf("%s %s\n", ui.Bold("Target:"), strings.Join(targetTools, ", "))
	if addDryRun {
		fmt.Println(ui.Yellow("(dry run — no files will be written)"))
	}
	fmt.Println()

	rb := installer.NewRollback()
	// Track files per top-level requested name, per tool
	type fileKey struct {
		name string
		tool string
	}
	filesByKey := make(map[fileKey][]string)

	for _, itemName := range items {
		item := reg.FindItem(itemName)

		for _, tool := range targetTools {
			// Skip items that don't support this tool
			if !item.Tool.Contains(tool) {
				continue
			}

			label := fmt.Sprintf("%s [%s]", ui.Cyan(itemName), tool)
			fmt.Printf("  %s (%s)\n", label, item.Type)

			files, err := installer.InstallItem(item, reg, registryDir, targetRoot, tool, addForce, addDryRun, rb)
			if err != nil {
				rb.Undo()
				ui.Die("Installation failed for '%s' (%s): %v", itemName, tool, err)
			}
			filesByKey[fileKey{itemName, tool}] = files
		}
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
				// Collect all files for this name and its deps across all target tools
				resolved, _ := reg.ResolveDeps(name)
				for _, tool := range targetTools {
					var allFiles []string
					for _, dep := range resolved {
						allFiles = append(allFiles, filesByKey[fileKey{dep, tool}]...)
					}
					if len(allFiles) > 0 || reg.FindItem(name).Tool.Contains(tool) {
						if err := tracker.Track(trackingPath, name, tool, sha, allFiles); err != nil {
							fmt.Printf("  %s could not update tracking for '%s' (%s): %v\n", ui.Yellow("warning:"), name, tool, err)
						}
					}
				}
			}
		}
	}

	fmt.Println()
	fmt.Println(ui.Green("Done."))
}

// sortedKeys returns sorted keys of a map.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func init() {
	addCmd.Flags().BoolVar(&addGlobal, "global", false, "Install to $HOME instead of current directory")
	addCmd.Flags().BoolVar(&addForce, "force", false, "Overwrite existing files")
	addCmd.Flags().BoolVar(&addDryRun, "dry-run", false, "Show what would be installed without writing")
	addCmd.Flags().StringVar(&addToolFlag, "tool", "", "Target tool (e.g. claude, opencode). Auto-detected if not set.")
	rootCmd.AddCommand(addCmd)
}
