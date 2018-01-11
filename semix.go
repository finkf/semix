package main

import (
	"fmt"
	"os"

	"bitbucket.org/fflo/semix/pkg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
