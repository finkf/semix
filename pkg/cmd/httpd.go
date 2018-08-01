package cmd

import (
	"gitlab.com/finkf/semix/pkg/httpd"
	"gitlab.com/finkf/semix/pkg/say"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
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
		&httpdHost,
		"host",
		"H",
		defaultHTTPDHost,
		"set host",
	)
}

func doHttpd(cmd *cobra.Command, args []string) error {
	setupSay()
	dir := args[0]
	s, err := httpd.New(
		httpd.WithHost(httpdHost),
		httpd.WithDaemon(daemonHost),
		httpd.WithDirectory(dir),
	)
	if err != nil {
		return errors.Wrapf(err, "cannot start httpd: dir: %s: host: %s: daemon: %s",
			dir, httpdHost, daemonHost)
	}
	say.Info("starting httpd on %s", httpdHost)
	return s.Start()
}
