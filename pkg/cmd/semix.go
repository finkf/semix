package cmd

import (
	"bitbucket.org/fflo/semix/pkg/rest"
	"github.com/spf13/cobra"
)

var (
	daemonHost string
)

var semixCmd = &cobra.Command{
	Use:   "semix",
	Long:  `SEMantic IndeXing`,
	Short: `SEMantic IndeXing`,
	Run:   semix,
}

func init() {
	semixCmd.PersistentFlags().StringVarP(
		&daemonHost,
		"daemon",
		"D",
		"localhost:6606",
		"set semix daemon host address",
	)
	semixCmd.AddCommand(putCmd)
	semixCmd.AddCommand(getCmd)
	semixCmd.AddCommand(searchCmd)
	semixCmd.AddCommand(infoCmd)
	semixCmd.AddCommand(daemonCmd)
}

func client() *rest.Client {
	client := rest.NewClient(daemonHost)
	return &client
}

func semix(cmd *cobra.Command, args []string) {

}

// Execute runs the main semix command.
func Execute() error {
	return semixCmd.Execute()
}
