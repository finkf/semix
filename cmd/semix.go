package cmd

import (
	"github.com/spf13/cobra"
)

var (
	daemonHost string
)

var semixCmd = &cobra.Command{
	Use:   "semix",
	Short: "semix is a semantic indexer",
	Run:   semix,
}

func init() {
	semixCmd.PersistentFlags().StringVarP(
		&daemonHost,
		"daemon",
		"d",
		"localhost:6606",
		"set semix daemon address",
	)
	semixCmd.AddCommand(putCmd)
}

func semix(cmd *cobra.Command, args []string) {

}

// Execute runs the main semix command.
func Execute() error {
	return semixCmd.Execute()
}
