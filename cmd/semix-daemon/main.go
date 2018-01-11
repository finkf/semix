package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/resource"
	"bitbucket.org/fflo/semix/pkg/rest"
)

var (
	dir     string
	host    string
	confg   string
	help    bool
	noCache bool
)

func init() {
	flag.StringVar(&dir, "dir",
		filepath.Join(os.Getenv("HOME"), "semix"), "set semix index directory")
	flag.StringVar(&host, "host", "localhost:6606", "set listen host")
	flag.StringVar(&confg, "resource", "testdata/topiczoom.toml", "set resource file")
	flag.BoolVar(&noCache, "no-cache", false, "do not load resource from cache")
	flag.BoolVar(&help, "help", false, "prints this help")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	s, err := server()
	if err != nil {
		log.Fatal(err)
	}
	run(s)
}

func run(s *rest.Server) {
	sigch := make(chan os.Signal)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigch
		log.Printf("got signale: %d", sig)
		if err := s.Close(); err != nil {
			log.Printf("error closing server: %s", err)
		}
	}()
	log.Printf("starting daemon on %s", host)
	err := s.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen and serve returned error: %s", err)
	}
}

func server() (*rest.Server, error) {
	index, err := index.New(dir, index.DefaultBufferSize)
	if err != nil {
		return nil, err
	}
	r, err := resource.Parse(confg, !noCache)
	if err != nil {
		return nil, err
	}
	return rest.New(host, dir, r, index)
}
