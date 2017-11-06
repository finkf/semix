package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rdfxml"
	"bitbucket.org/fflo/semix/pkg/rest"
	"bitbucket.org/fflo/semix/pkg/traits"
)

var (
	dir  string
	host string
	rdf  string
	help bool
)

func init() {
	flag.StringVar(&dir, "dir",
		filepath.Join(os.Getenv("HOME"), "semix"),
		"set semix index directory")
	flag.StringVar(&host, "host", "localhost:6060", "set listen host")
	flag.StringVar(&rdf, "rdf",
		filepath.Join(os.Getenv("HOME"),
			"/devel/priv/semix/misc/data/topiczoom.skos.rdf.xml"),
		"set RDF input file")
	flag.BoolVar(&help, "help", false, "prints this help")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	index, err := index.New(
		dir,
		index.WithBufferSize(5),
	)
	if err != nil {
		log.Fatal(err)
	}
	is, err := os.Open(rdf)
	if err != nil {
		log.Fatal(err)
	}
	defer is.Close()
	t := traits.New(
		traits.WithIgnoreURLs(
			"http://www.w3.org/2004/02/skos/core#narrower",
		),
		traits.WithTransitiveURLs(
			"http://www.w3.org/2004/02/skos/core#broader",
			"http://www.w3.org/2004/02/skos/core#narrower",
		),
		traits.WithNameURLs(
			"http://www.w3.org/2004/02/skos/core#prefLabel",
		),
		traits.WithDistinctURLs(
			"http://www.w3.org/2004/02/skos/core#altLabel",
		),
	)
	parser := rdfxml.NewParser(is)
	s, err := rest.New(host, parser, t, index)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("starting server")
	log.Fatal(s.ListenAndServe())
}
