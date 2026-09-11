// Package application coordinates Calathea commands and queries.
//
// Application code may depend on domain and deterministic services plus
// application-owned ports. Domain and deterministic packages must not depend
// back on this package or on infrastructure adapters.
package application

import (
	"fmt"
	"io"
)

// Version identifies the executable build. Release packaging overrides the
// development default with -ldflags -X.
var Version = "dev"

// Run executes the minimal CLI application boundary and returns a process exit code.
// Feature commands are introduced by later roadmap issues; this skeleton exists to
// prove dependency direction and local executable packaging.
func Run(stdout, stderr io.Writer, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(stdout, "calathea: local portfolio orientation")
		return 0
	}

	switch args[0] {
	case "version":
		fmt.Fprintf(stdout, "calathea %s\n", Version)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		return 2
	}
}
