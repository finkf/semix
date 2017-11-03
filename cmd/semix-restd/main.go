package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rdfxml"
	"bitbucket.org/fflo/semix/pkg/semix"
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
	log.Printf("reading RDF-XML")
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
	graph, dictionary, err := semix.Parse(parser, t)
	if err != nil {
		log.Fatal(err)
	}
	dfa := semix.NewDFA(dictionary, graph)
	log.Printf("done reading RDF-XML")
	log.Printf("starting the server")
	h := handle{dfa: dfa, g: graph, d: dictionary, i: index}
	http.HandleFunc("/search", requestFunc(h.search))
	http.HandleFunc("/put", requestFunc(h.put))
	http.HandleFunc("/get", requestFunc(h.get))
	http.HandleFunc("/ctx", requestFunc(h.ctx))
	http.HandleFunc("/info", requestFunc(h.info))
	log.Fatalf(http.ListenAndServe(host, nil).Error())
}
