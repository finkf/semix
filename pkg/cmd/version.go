package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"bitbucket.org/fflo/semix/pkg/say"
	"github.com/spf13/cobra"
)

// Semantic version variables.
// These constants are automatically updated.
// DO NOT EDIT BY HAND.
const (
	major = 0
	minor = 1
	patch = 1
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
	if jsonOutput {
		return printJSONVersion()
	}
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
	fmt.Printf("%s\n", versionString())
	return nil
}

func printJSONVersion() error {
	return json.NewEncoder(os.Stdout).Encode(
		struct {
			Major, Minor, Patch int
			Version             string
		}{major, minor, patch, versionString()})
}

func versionString() string {
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}
