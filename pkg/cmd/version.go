package cmd

import (
	"fmt"

	"bitbucket.org/fflo/semix/pkg/say"
	"github.com/spf13/cobra"
)

// Semantic version variables.
// These constants are automatically updated.
// DO NOT EDIT BY HAND.
const (
	major = 0
	minor = 0
	patch = 0
)

var (
	pmajor, pminor, ppatch bool
)

func init() {
	versionCmd.Flags().BoolVarP(&pmajor, "major", "M", false, "print major version")
	versionCmd.Flags().BoolVarP(&pminor, "minor", "m", false, "print minor version")
	versionCmd.Flags().BoolVarP(&ppatch, "patch", "p", false, "print patch version")
}

var versionCmd = &cobra.Command{
	Use:          "version [options]",
	Short:        "Get semantic version of semix",
	Long:         "Get semantic version of semix",
	RunE:         version,
	SilenceUsage: true,
}

func version(cmd *cobra.Command, args []string) error {
	say.SetDebug(debug)
	if pmajor {
		fmt.Printf("%d\n", major)
		return nil
	}
	if pminor {
		fmt.Printf("%d\n", minor)
		return nil
	}
	if ppatch {
		fmt.Printf("%d\n", patch)
		return nil
	}
	fmt.Printf("v%d.%d.%d\n", major, minor, patch)
	return nil
}
