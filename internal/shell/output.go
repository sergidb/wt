package shell

import "fmt"

// CdPrefix is printed before a path when the shell wrapper should cd.
const CdPrefix = "__WT_CD__:"

// PrintCdPath writes a cd-signal line to stdout.
func PrintCdPath(path string) {
	fmt.Print(CdPrefix + path)
}

// PrintDisplay writes normal display output to stdout.
func PrintDisplay(s string) {
	fmt.Print(s)
}
