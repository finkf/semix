package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rest"
	x "bitbucket.org/fflo/semix/pkg/semix"
	"github.com/spf13/cobra"
)

var httpdCmd = &cobra.Command{
	Use:          "httpd",
	Short:        "Start an http server",
	Long:         "The httpd command starts an http server.",
	RunE:         httpd,
	SilenceUsage: true,
}

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


func httpd(cmd *cobra.Command, args []string) error {
	s, err := httpd.New(
		httpd.WithHost(host),
		httpd.WithDaemon(daemonHost),
		httpd.WithDirectory(dir),
	)
	if err != nil {
		return errors.Wrapf(err, "could not start httpd: dir: %s, host: % daemon: %s",
			dir, host daemonHost);
	}
	return s.Start()
}
