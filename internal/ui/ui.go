package ui

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	Bold   = color.New(color.Bold).SprintFunc()
	Dim    = color.New(color.Faint).SprintFunc()
	Cyan   = color.New(color.FgCyan).SprintFunc()
	Green  = color.New(color.FgGreen).SprintFunc()
	Yellow = color.New(color.FgYellow).SprintFunc()
	Red    = color.New(color.FgRed).SprintFunc()
)

// Die prints an error message and exits.
func Die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s %s\n", Red("error:"), fmt.Sprintf(format, args...))
	os.Exit(1)
}
