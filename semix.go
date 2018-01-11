package main

import (
	"os"

	"bitbucket.org/fflo/semix/pkg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// no need to print error
		// since cobra takes care of this
		// fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
