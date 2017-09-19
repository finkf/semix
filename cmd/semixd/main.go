package main

import (
	"fmt"
	"net/http"
	"os"

	"bitbucket.org/fflo/semix/pkg/rdfxml"
	"bitbucket.org/fflo/semix/pkg/semix"
	"github.com/sirupsen/logrus"
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

var file = "/home/flo/devel/priv/semix/misc/data/topiczoom.skos.rdf.xml"

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Infof("reading RDF-XML")
	p := rdfxml.NewParser()
	if err := p.ParseFile(file); err != nil {
		logrus.Fatal(err)
	}
	g, d := p.Get()
	i := semix.NewDirIndex(semixdir, 10)
	dfa := semix.NewDFA(d, g)
	logrus.Infof("done reading RDF-XML")
	logrus.Infof("starting the server")
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		search(g, d, w, r)
	})
	http.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		put(dfa, i, w, r)
	})
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		get(dfa, i, w, r)
	})
	logrus.Fatalf(http.ListenAndServe(":6060", nil).Error())
}
