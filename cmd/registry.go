package cmd

import (
	"fmt"
	"os"

	"github.com/bensch98/arcane/internal/git"
	"github.com/bensch98/arcane/internal/registry"
	"github.com/bensch98/arcane/internal/ui"
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage the item registry",
}

var registryFetchCmd = &cobra.Command{
	Use:   "fetch [url]",
	Short: "Fetch or re-fetch the registry from a remote URL",
	Long: `Fetch the registry from a remote git URL into the local cache.

If no URL is given, the default registry is used:
  ` + registry.DefaultRegistryURL + `

Examples:
  arcane registry fetch
  arcane registry fetch https://github.com/myorg/my-registry.git`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip default registry loading.
	},
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := registry.DefaultRegistryURL
		if len(args) == 1 {
			url = args[0]
		}

		cacheDir := registry.CacheDir()

		// If cache dir already exists, remove it for a clean fetch
		if _, err := os.Stat(cacheDir); err == nil {
			fmt.Printf("Removing existing cache at %s\n", ui.Dim(cacheDir))
			if err := os.RemoveAll(cacheDir); err != nil {
				ui.Die("Cannot remove existing cache: %v", err)
			}
		}

		fmt.Printf("Fetching registry from %s ...\n", ui.Cyan(url))

		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			ui.Die("Cannot create cache directory: %v", err)
		}
		// Remove the dir so git clone can create it
		os.Remove(cacheDir)

		if err := git.Clone(url, cacheDir); err != nil {
			ui.Die("Clone failed: %v", err)
		}

		fmt.Println(ui.Green("✓ Registry fetched to " + cacheDir))

		// Show a quick summary
		reg, err := registry.Load(cacheDir + "/registry.json")
		if err != nil {
			return
		}
		fmt.Printf("\n  %d items available. Run %s to browse.\n", len(reg.Items), ui.Cyan("arcane list"))
	},
}

var registryStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show registry location and status",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip default registry loading.
	},
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := registry.FindRegistryDir()
		if err != nil {
			fmt.Println(ui.Yellow("No registry found."))
			fmt.Printf("Run %s to fetch the default registry.\n", ui.Cyan("arcane registry fetch"))
			return
		}

		fmt.Printf("Registry: %s\n", ui.Cyan(dir))

		if git.IsRepo(dir) {
			sha := git.RevParseShort(dir)
			if sha != "" {
				fmt.Printf("Revision: %s\n", sha)
			}
		}

		reg, err := registry.Load(dir + "/registry.json")
		if err != nil {
			fmt.Printf("Status:   %s\n", ui.Red("invalid registry.json"))
			return
		}
		fmt.Printf("Items:    %d\n", len(reg.Items))
	},
}

func init() {
	registryCmd.AddCommand(registryFetchCmd)
	registryCmd.AddCommand(registryStatusCmd)
	rootCmd.AddCommand(registryCmd)
}
