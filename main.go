package main

import (
	"os"

	"github.com/sergidb/wt/cmd"
)

func main() {
	// Force color output — stdout may be captured by shell wrapper,
	// which makes lipgloss think the terminal has no color support.
	os.Setenv("CLICOLOR_FORCE", "1")

	cmd.Execute()
}
