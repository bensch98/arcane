package cmd

import (
	"os"

	"github.com/bensch98/arcane/internal/registry"
	"github.com/bensch98/arcane/internal/ui"
	"github.com/spf13/cobra"
)

var (
	reg         *registry.Registry
	registryDir string
)

var rootCmd = &cobra.Command{
	Use:   "arcane",
	Short: "Agentic Registry CLI",
	Long:  "arcane — A shadcn-style registry for Claude Code commands, scripts, skills, and hooks.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip registry loading for init (doesn't need it)
		if cmd.Name() == "init" {
			return
		}
		var err error
		registryDir, err = registry.FindRegistryDir()
		if err != nil {
			ui.Die("%v", err)
		}
		reg, err = registry.Load(registryDir + "/registry.json")
		if err != nil {
			ui.Die("%v", err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
