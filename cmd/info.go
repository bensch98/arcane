package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bensch98/arcane/internal/ui"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show item details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		item := reg.FindItem(name)
		if item == nil {
			ui.Die("Item '%s' not found.", name)
		}

		fmt.Println(ui.Bold(item.Name))
		fmt.Println(ui.Dim(item.Description))
		fmt.Println()
		fmt.Printf("  Tool:         %s\n", item.Tool.String())
		fmt.Printf("  Type:         %s\n", item.Type)

		if len(item.Tags) > 0 {
			fmt.Printf("  Tags:         %s\n", strings.Join(item.Tags, ", "))
		}
		if len(item.Dependencies) > 0 {
			fmt.Printf("  Dependencies: %s\n", strings.Join(item.Dependencies, ", "))
		}

		fmt.Println("  Files:")
		for _, f := range item.Files {
			toolLabel := ""
			if f.Tool != "" {
				toolLabel = fmt.Sprintf(" [%s]", f.Tool)
			}
			if f.Target != "" {
				fmt.Printf("    %s → %s%s\n", f.Src, f.Target, toolLabel)
			} else {
				fmt.Printf("    %s%s\n", f.Src, toolLabel)
			}
		}

		// File preview for single-file items (or first file with content)
		if len(item.Files) >= 1 {
			fullPath := filepath.Join(registryDir, item.Files[0].Src)
			if f, err := os.Open(fullPath); err == nil {
				defer f.Close()
				fmt.Println()
				fmt.Println(ui.Dim("--- preview (first 20 lines) ---"))
				scanner := bufio.NewScanner(f)
				lineCount := 0
				totalLines := 0
				for scanner.Scan() {
					totalLines++
					if lineCount < 20 {
						fmt.Println(scanner.Text())
						lineCount++
					}
				}
				if totalLines > 20 {
					fmt.Println(ui.Dim(fmt.Sprintf("... (%d lines total)", totalLines)))
				}
			}
		}

		// Hook info (Claude Code)
		if item.Type == "hook" && item.HookMerge != nil {
			fmt.Println()
			fmt.Printf("  Hook event:   %s\n", item.HookMerge.Event)
			if matcher, ok := item.HookMerge.Entry["matcher"].(string); ok && matcher != "" {
				fmt.Printf("  Matcher:      %s\n", matcher)
			}
		}

		// Config merge info (OpenCode)
		if (item.Type == "formatter" || item.Type == "config-merge") && item.ConfigMerge != nil {
			fmt.Println()
			fmt.Printf("  Config path:  %s\n", item.ConfigMerge.Path)
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
