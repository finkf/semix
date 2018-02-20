// Package say implements a simple threadsave logger
// that exposes Info and optional Debug type messages.
// It simply uses go's default logger.
package say

import (
	"fmt"
	"io"
	"log"
	"sync/atomic"
)

func init() {
	// set default logging flags
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
}

var (
	debug int32
)

const (
	dmsg = "[DEBG]"
	imsg = "[INFO]"
)

// SetDebug sets debugging output.
func SetDebug(f bool) {
	if f {
		atomic.StoreInt32(&debug, 1)
	} else {
		atomic.StoreInt32(&debug, 0)
	}
}

// SetOutput sets the output of the writer.
func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

// Info logs an info message.
func Info(msg string, args ...interface{}) {
	log.Printf("%s %s", imsg, fmt.Sprintf(msg, args))
}

// Debug logs an debug message if debugging is enabled
func Debug(msg string, args ...interface{}) {
	if atomic.LoadInt32(&debug) == 0 {
		return
	}
	log.Printf("%s %s", dmsg, fmt.Sprintf(msg, args))
}
