package shell

import (
	"fmt"
	"io"
)

// CdPrefix is printed before a path when the shell wrapper should cd.
const CdPrefix = "__WT_CD__:"

// PrintCdPath writes a cd-signal line to the given writer.
func PrintCdPath(w io.Writer, path string) {
	fmt.Fprint(w, CdPrefix+path)
}

// PrintDisplay writes normal display output to the given writer.
func PrintDisplay(w io.Writer, s string) {
	fmt.Fprint(w, s)
}
