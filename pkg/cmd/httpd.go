package cmd

import (
	"os"

	"bitbucket.org/fflo/semix/pkg/httpd"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	dir      string
	host     string
	httpdCmd = &cobra.Command{
		Use:          "httpd",
		Short:        "Start an http server",
		Long:         "The httpd command starts an http server.",
		RunE:         doHttpd,
		SilenceUsage: true,
	}
)

func init() {
	httpdCmd.Flags().StringVarP(
		&dir,
		"dir",
		"d",
		os.Getenv("SEMIX_HTTPD_DIR"),
		"set template directory",
	)
	httpdCmd.Flags().StringVarP(
		&host,
		"host",
		"H",
		"localhost:8080",
		"set host",
	)
}

func doHttpd(cmd *cobra.Command, args []string) error {
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
