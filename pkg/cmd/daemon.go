package cmd

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/resource"
	"bitbucket.org/fflo/semix/pkg/rest"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:          "daemon",
	Short:        "daemon [options...]",
	Long:         `The daemon command starts the semix daemon.`,
	RunE:         daemon,
	SilenceUsage: true,
}

var (
	daemonDir       string
	daemonResource  string
	daemonNoCache   bool
	indexBufferSize int
)

func init() {
	daemonCmd.Flags().StringVarP(&daemonDir, "dir", "d",
		filepath.Join(os.Getenv("HOME"), "semix"), "set semix index directory")
	daemonCmd.Flags().StringVarP(&daemonResource, "resource", "r",
		"semix.toml", "set path of resource file")
	daemonCmd.Flags().BoolVar(&daemonNoCache, "no-cache",
		false, "do not load cached resources")
	daemonCmd.Flags().IntVar(&indexBufferSize, "index-size",
		index.DefaultBufferSize, "set buffer size of index")
}

func daemon(cmd *cobra.Command, args []string) error {
	s, err := newServer()
	if err != nil {
		return err
	}
	sigch := make(chan os.Signal)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigch
		log.Printf("got signale: %d", sig)
		if err := s.Close(); err != nil {
			log.Printf("error closing server: %s", err)
		}
	}()
	log.Printf("starting daemon on %s", daemonHost)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func newServer() (*rest.Server, error) {
	index, err := index.New(daemonDir, indexBufferSize)
	if err != nil {
		return nil, err
	}
	r, err := resource.Parse(daemonResource, !daemonNoCache)
	if err != nil {
		return nil, err
	}
	return rest.New(daemonHost, daemonDir, r, index)
}
