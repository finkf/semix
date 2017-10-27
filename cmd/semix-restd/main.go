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
)

var (
	semixdir string
	host     string
	rdf      string
	help     bool
)

func init() {
	flag.StringVar(&semixdir, "semixdir",
		filepath.Join(os.Getenv("HOME"), "semix"),
		"set semix index directory")
	flag.StringVar(&host, "host", "localhost:6060", "set listen host")
	flag.StringVar(&rdf, "rdf",
		filepath.Join(os.Getenv("HOME"), "/devel/priv/semix/misc/data/topiczoom.skos.rdf.xml"),
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
		semixdir,
		index.WithBufferSize(5),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("reading RDF-XML")
	p := rdfxml.NewParser(
		rdfxml.WithIgnoreURLs(
			"http://www.w3.org/2004/02/skos/core#narrower",
		),
		rdfxml.WithTransitiveURLs(
			"http://www.w3.org/2004/02/skos/core#broader",
			"http://www.w3.org/2004/02/skos/core#narrower",
		),
	)
	if err := p.ParseFile(rdf); err != nil {
		log.Fatal(err)
	}
	graph, dictionary := p.Get()
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
	log.Fatalf(http.ListenAndServe(host, nil).Error())
}
