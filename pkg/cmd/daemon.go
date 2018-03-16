package cmd

import (
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/resource"
	"bitbucket.org/fflo/semix/pkg/rest"
	"bitbucket.org/fflo/semix/pkg/say"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:          "daemon <ressource>",
	Short:        "Starts the daemon",
	Long:         `The daemon command starts the semix daemon.`,
	RunE:         daemon,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
}

var (
	daemonDir       string
	daemonNoCache   bool
	indexBufferSize int
)

func semixDir() string {
	res := os.Getenv("SEMIXPATH")
	if res == "" {
		res = filepath.Join(os.Getenv("HOME"), "semix")
	}
	return res
}

func init() {
	daemonCmd.Flags().StringVarP(&daemonDir, "dir", "d",
		semixDir(), "set semix index directory")
	daemonCmd.Flags().BoolVar(&daemonNoCache, "no-cache",
		false, "do not load cached resources")
	daemonCmd.Flags().IntVar(&indexBufferSize, "index-size",
		index.DefaultBufferSize, "set buffer size of index")
}

func daemon(cmd *cobra.Command, args []string) error {
	setupSay()
	s, err := newServer(args[0])
	if err != nil {
		return err
	}
	sigch := make(chan os.Signal)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigch
		say.Info("got signal: %s", sig)
		if err := s.Close(); err != nil {
			say.Info("error closing server: %s", err)
		}
	}()
	say.Info("starting daemon on %s", daemonHost)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func newServer(res string) (*rest.Server, error) {
	index, err := index.NewDir(daemonDir, indexBufferSize)
	if err != nil {
		return nil, err
	}
	r, err := resource.Parse(res, !daemonNoCache)
	if err != nil {
		return nil, err
	}
	return rest.New(daemonHost, daemonDir, r, index)
}
