package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rdfxml"
	"bitbucket.org/fflo/semix/pkg/rest"
	"bitbucket.org/fflo/semix/pkg/semix"
	"bitbucket.org/fflo/semix/pkg/traits"
	"bitbucket.org/fflo/semix/pkg/turtle"
	"github.com/BurntSushi/toml"
)

var (
	dir   string
	host  string
	confg string
	help  bool
)

func init() {
	flag.StringVar(&dir, "dir",
		filepath.Join(os.Getenv("HOME"), "semix"),
		"set semix index directory")
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
	index, err := index.New(dir)
	panicIf(err)
	config, err := readConfig(confg)
	panicIf(err)
	log.Printf("config: %v", config)
	is, err := os.Open(config.Parser.File)
	panicIf(err)
	defer is.Close()
	parser, err := newParser(is, config)
	panicIf(err)
	s, err := rest.New(host, parser, config.traits(), index)
	panicIf(err)
	log.Printf("starting the server")
	log.Fatal(s.ListenAndServe())
}

func panicIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type parser struct {
	File, Type string
}

type urls struct {
	Ignore     []string
	Transitive []string
	Symmetric  []string
	Name       []string
	Distinct   []string
	Ambiguous  []string
}

type config struct {
	Parser parser
	URLs   urls
}

func (c config) traits() semix.Traits {
	return traits.New(
		traits.WithIgnoreURLs(c.URLs.Ignore...),
		traits.WithTransitiveURLs(c.URLs.Transitive...),
		traits.WithNameURLs(c.URLs.Name...),
		traits.WithAmbiguousURLs(c.URLs.Ambiguous...),
		traits.WithDistinctURLs(c.URLs.Distinct...),
	)
}

func readConfig(file string) (config, error) {
	var c config
	if _, err := toml.DecodeFile(file, &c); err != nil {
		return config{}, err
	}
	return c, nil
}

func newParser(r io.Reader, c config) (semix.Parser, error) {
	switch strings.ToLower(c.Parser.Type) {
	case "rdfxml":
		return rdfxml.NewParser(r), nil
	case "turtle":
		return turtle.NewParser(r), nil
	default:
		return nil, fmt.Errorf("invalid type: %s", c.Parser.Type)
	}
}
