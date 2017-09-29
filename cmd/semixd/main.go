package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rdfxml"
	"bitbucket.org/fflo/semix/pkg/semix"
)

var semixdir string

// check for environment variable SEMIXDIR
func init() {
	semixdir = os.Getenv("SEMIXDIR")
	if semixdir == "" {
		panic("environment variable SEMIXDIR not set")
	}
	info, err := os.Lstat(semixdir)
	if err != nil {
		panic(fmt.Sprintf("could not stat %s: %v", semixdir, err))
	}
	if !info.IsDir() {
		panic(fmt.Sprintf("%s: not a directory", semixdir))
	}
}

var file = os.Getenv("HOME") + "/devel/priv/semix/misc/data/topiczoom.skos.rdf.xml"

func main() {
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
	if err := p.ParseFile(file); err != nil {
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
	log.Fatalf(http.ListenAndServe(":6060", nil).Error())
}
