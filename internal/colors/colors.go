package colors

// Names is the canonical list of available service colors.
var Names = []string{"green", "cyan", "yellow", "magenta", "blue", "red", "white"}

// ANSICodes maps color names to ANSI escape sequences.
var ANSICodes = map[string]string{
	"red":     "\033[31m",
	"green":   "\033[32m",
	"yellow":  "\033[33m",
	"blue":    "\033[34m",
	"magenta": "\033[35m",
	"cyan":    "\033[36m",
	"white":   "\033[37m",
}

// ANSIReset is the ANSI escape sequence to reset color.
const ANSIReset = "\033[0m"
