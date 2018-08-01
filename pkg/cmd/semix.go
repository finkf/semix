package cmd

import (
	"os"
	"strings"

	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"gitlab.com/finkf/semix/pkg/say"
)

const (
	// defaultDaemonHost specifies the default
	// host for the semix daemon.
	defaultDaemonHost = "localhost:6060"
	// defaultHTTPDHost specifies the default
	// host for the HTTP server.
	defaultHTTPDHost = "localhost:8080"
)

var (
	daemonHost string
	httpdHost  string
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
		defaultDaemonHost,
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

func setupSay() {
	say.SetDebug(debug)
	// set color mode true if connected to a terminal
	say.SetColor(isatty.IsTerminal(os.Stdout.Fd()))
	// if nocolor isgiven disable colors no matter what
	if nocolor {
		say.SetColor(false)
	}
}

// DaemonHost returns the configured address for the daemon
func DaemonHost() string {
	return hostWithPrefix(daemonHost)
}

// HTTPDHost returns the configured address for the HTTP-daemon
func HTTPDHost() string {
	return hostWithPrefix(httpdHost)
}

func hostWithPrefix(host string) string {
	if !strings.HasPrefix(host, "http://") || !strings.HasPrefix(host, "https://") {
		return "http://" + host
	}
	return host
}

// Execute runs the main semix command.
func Execute() error {
	return semixCmd.Execute()
}
