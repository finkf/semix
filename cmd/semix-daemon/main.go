package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"bitbucket.org/fflo/semix/pkg/config"
	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rest"
)

var (
	dir   string
	host  string
	confg string
	help  bool
)

func init() {
	flag.StringVar(&dir, "dir",
		filepath.Join(os.Getenv("HOME"), "semix"), "set semix index directory")
	flag.StringVar(&host, "host", "localhost:6060", "set listen host")
	flag.StringVar(&confg, "config", "testdata/topiczoom.toml", "set configuration file")
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
		select {
		case sig := <-sigch:
			log.Printf("got signal: %d", sig)
			if err := s.Close(); err != nil {
				log.Fatalf("could not close server: %s", err)
			}
		}
	}()
	log.Printf("starting the server on %s", host)
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
	g, d, err := config.Parse(confg)
	if err != nil {
		return nil, err
	}
	return rest.New(host, dir, g, d, index), nil
}
