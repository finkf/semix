package main

import (
	"os"

	"gitlab.com/finkf/semix/pkg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// no need to print error message
		// since cobra takes care of this
		os.Exit(1)
	}
}
