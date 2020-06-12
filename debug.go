package gig

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	// DefaultWriter is the default io.Writer used by gig for debug output and
	// middleware output like Logger() or Recovery().
	// Note that both Logger and Recovery provides custom ways to configure their
	// output io.Writer.
	// To support coloring in Windows use:
	// 		import "github.com/mattn/go-colorable"
	// 		gin.DefaultWriter = colorable.NewColorableStdout()
	DefaultWriter io.Writer = os.Stdout

	// Debug enables gig to print its internal debug messages.
	Debug = true
)

func debugPrint(format string, values ...interface{}) {
	if Debug {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(DefaultWriter, "[gig-debug] "+format, values...)
	}
}
