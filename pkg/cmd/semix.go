package cmd

import (
	"os"
	"strings"

	"bitbucket.org/fflo/semix/pkg/client"
	"bitbucket.org/fflo/semix/pkg/say"
	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

const (
	// DaemonHost specifies the default  host for the semix daemon.
	DaemonHost = "localhost:6060"
	// HTTPDHost specifies the default host for the HTTP server.
	HTTPDHost = "localhost:8080"
)

var (
	daemonHost string
	jsonOutput bool
	debug      bool
	nocolor    bool
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
		DaemonHost,
		"set semix daemon host address",
	)
	semixCmd.PersistentFlags().BoolVarP(
		&jsonOutput,
		"json",
		"J",
		false,
		"set JSON output",
	)
	semixCmd.PersistentFlags().BoolVarP(
		&debug,
		"verbose",
		"V",
		false,
		"enable debugging output",
	)
	semixCmd.PersistentFlags().BoolVarP(
		&nocolor,
		"no-colors",
		"N",
		false,
		"disable colors in log messages",
	)
	semixCmd.AddCommand(versionCmd)
	semixCmd.AddCommand(putCmd)
	semixCmd.AddCommand(getCmd)
	semixCmd.AddCommand(searchCmd)
	semixCmd.AddCommand(infoCmd)
	semixCmd.AddCommand(daemonCmd)
	semixCmd.AddCommand(httpdCmd)
}

func newClient(opts ...client.Option) *client.Client {
	host := daemonHost
	if !strings.HasPrefix(host, "http://") ||
		!strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	client := client.New(host, opts...)
	return client
}

func setupSay() {
	say.SetDebug(debug)
	// set color mode true if connected to a terminal
	say.SetColor(isatty.IsTerminal(os.Stdout.Fd()))
	// if nocolor isgiven disable colors no matter what
	if nocolor {
		say.SetColor(false)
	}
}

// Execute runs the main semix command.
func Execute() error {
	return semixCmd.Execute()
}
