package cmd

import (
	"strings"

	"bitbucket.org/fflo/semix/pkg/rest"
	"github.com/spf13/cobra"
)

var (
	daemonHost string
	jsonOutput bool
)

var semixCmd = &cobra.Command{
	Use:   "semix",
	Long:  `SEMantic IndeXing`,
	Short: `SEMantic IndeXing`,
}

func init() {
	semixCmd.PersistentFlags().StringVarP(
		&daemonHost,
		"daemon",
		"D",
		"localhost:6606",
		"set semix daemon host address",
	)
	semixCmd.PersistentFlags().BoolVarP(
		&jsonOutput,
		"json",
		"J",
		false,
		"set JSON output",
	)
	semixCmd.AddCommand(putCmd)
	semixCmd.AddCommand(getCmd)
	semixCmd.AddCommand(searchCmd)
	semixCmd.AddCommand(infoCmd)
	semixCmd.AddCommand(daemonCmd)
	semixCmd.AddCommand(httpdCmd)
}

func newClient() *rest.Client {
	host := daemonHost
	if !strings.HasPrefix(host, "http://") ||
		!strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	client := rest.NewClient(host)
	return &client
}

// Execute runs the main semix command.
func Execute() error {
	return semixCmd.Execute()
}