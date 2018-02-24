// Package say implements a simple threadsave logger
// that exposes Info and optional Debug type messages.
// It simply uses go's default logger.
package say

import (
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
)

func init() {
	// set default logging flags
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
}

var (
	debug int32
	mutex sync.Mutex
	dmsg  = "[DBUG]"
	imsg  = "[INFO]"
	red   = color.New(color.FgRed).SprintFunc()
	green = color.New(color.FgGreen).SprintFunc()
)

// SetDebug sets debugging output.
func SetDebug(f bool) {
	if f {
		atomic.StoreInt32(&debug, 1)
	} else {
		atomic.StoreInt32(&debug, 0)
	}
}

// SetColor sets debugging output.
func SetColor(f bool) {
	mutex.Lock()
	defer mutex.Unlock()
	if f {
		dmsg = "[" + red("DBUG") + "]"
		imsg = "[" + green("INFO") + "]"
	} else {
		dmsg = "[DBUG]"
		imsg = "[INFO]"
	}
}

// SetOutput sets the output of the writer.
func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

// Info logs an info message.
func Info(msg string, args ...interface{}) {
	log.Printf("%s %s", imsg, fmt.Sprintf(msg, args...))
}

// Debug logs an debug message if debugging is enabled
func Debug(msg string, args ...interface{}) {
	if atomic.LoadInt32(&debug) == 0 {
		return
	}
	log.Printf("%s %s", dmsg, fmt.Sprintf(msg, args...))
}
