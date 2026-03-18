package cmd

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/bensch98/arcane/internal/ui"
	"github.com/spf13/cobra"
)

var listToolFlag string
var listTypeFlag string

var listCmd = &cobra.Command{
	Use:   "list [SEARCH]",
	Short: "Browse items",
	Long:  "List registry items, optionally filtered by tool, type, or search term.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		search := ""
		if len(args) > 0 {
			search = args[0]
		}

		type entry struct {
			typ, name, desc, tools string
		}
		var entries []entry

		for _, item := range reg.Items {
			if listToolFlag != "" && !item.Tool.Contains(listToolFlag) {
				continue
			}
			if listTypeFlag != "" && item.Type != listTypeFlag {
				continue
			}
			if search != "" {
				re, err := regexp.Compile("(?i)" + search)
				if err != nil {
					ui.Die("Invalid search pattern: %v", err)
				}
				tags := strings.Join(item.Tags, " ")
				if !re.MatchString(item.Name) && !re.MatchString(item.Description) && !re.MatchString(tags) {
					continue
				}
			}
			entries = append(entries, entry{item.Type, item.Name, item.Description, item.Tool.String()})
		}

		if len(entries) == 0 {
			fmt.Println("No items found.")
			return
		}

		sort.Slice(entries, func(i, j int) bool {
			if entries[i].typ != entries[j].typ {
				return entries[i].typ < entries[j].typ
			}
			return entries[i].name < entries[j].name
		})

		currentType := ""
		for _, e := range entries {
			if e.typ != currentType {
				if currentType != "" {
					fmt.Println()
				}
				fmt.Println(ui.Bold(e.typ + "s"))
				currentType = e.typ
			}
			toolBadge := ui.Dim("[" + e.tools + "]")
			fmt.Printf("  %-30s %s %s\n", ui.Cyan(e.name), toolBadge, e.desc)
		}
	},
}

func init() {
	listCmd.Flags().StringVar(&listToolFlag, "tool", "", "Filter by tool (e.g. claude, opencode)")
	listCmd.Flags().StringVar(&listTypeFlag, "type", "", "Filter by type (e.g. command, script, hook, formatter, plugin)")
	rootCmd.AddCommand(listCmd)
}
