package cmd

import (
	"bitbucket.org/fflo/semix/pkg/httpd"
	"bitbucket.org/fflo/semix/pkg/say"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	host     string
	httpdCmd = &cobra.Command{
		Use:          "httpd directory",
		Short:        "Start an http server",
		Long:         "The httpd command starts an http server.",
		RunE:         doHttpd,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
	}
)

func init() {
	httpdCmd.Flags().StringVarP(
		&host,
		"host",
		"H",
		"localhost:80",
		"set host",
	)
}

func doHttpd(cmd *cobra.Command, args []string) error {
	say.SetDebug(debug)
	dir := args[0]
	s, err := httpd.New(
		httpd.WithHost(host),
		httpd.WithDaemon(daemonHost),
		httpd.WithDirectory(dir),
	)
	if err != nil {
		return errors.Wrapf(err, "could not start httpd: dir: %s: host: %s: daemon: %s",
			dir, host, daemonHost)
	}
	return s.Start()
}
