package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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
		index(dfa, i, w, r)
	})
	http.ListenAndServe(":8080", nil)
}

type Link struct {
	Predicate, Object string
}

type LookupInfo struct {
	Query   string
	Subject string
	Links   []Link
	Entries []string
}

func search(g *semix.Graph, d map[string]*semix.Concept, w http.ResponseWriter, r *http.Request) {
	logrus.Infof("serving request for %s", r.RequestURI)
	if r.Method != "GET" {
		logrus.Infof("invalid method: %s", r.Method)
		http.Error(w, "not a GET request", http.StatusBadRequest)
		return
	}
	q := r.URL.Query()["q"]
	if len(q) != 1 {
		logrus.Infof("invalid query: %v", q)
		http.Error(w, "invalid query paramters", http.StatusBadRequest)
		return
	}
	c := lookup(d, q[0])
	if c == nil {
		logrus.Infof("could not find %q", q[0])
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	info := makeLookupInfo(d, c)
	info.Query = q[0]
	e := json.NewEncoder(w)
	if err := e.Encode(&info); err != nil {
		logrus.Infof("could not encode json: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	logrus.Infof("handled %s", r.URL.Path)
}

func makeLookupInfo(d map[string]*semix.Concept, c *semix.Concept) LookupInfo {
	info := LookupInfo{Subject: c.URL()}
	for str, cc := range d {
		if cc == c {
			info.Entries = append(info.Entries, str)
		}
	}
	c.Edges(func(edge semix.Edge) {
		info.Links = append(info.Links, Link{
			Predicate: edge.P.URL(),
			Object:    edge.O.URL(),
		})
	})
	return info
}

func lookup(d map[string]*semix.Concept, q string) *semix.Concept {
	for _, c := range d {
		if strings.Contains(c.URL(), q) {
			return c
		}
	}
	return nil
}

type TokenInfo struct {
	Token, ConceptURL string
	Begin, End        int
}

type IndexInfo struct {
	Tokens []TokenInfo
}

func index(dfa semix.DFA, i semix.Index, w http.ResponseWriter, r *http.Request) {
	logrus.Infof("serving request for %s", r.RequestURI)
	if r.Method != "POST" {
		logrus.Infof("invalid method: %s", r.Method)
		http.Error(w, "not a POST request", http.StatusBadRequest)
		return
	}
	stream, cancel := makeStream(dfa, r.Body)
	defer cancel()
	info := IndexInfo{Tokens: []TokenInfo{}} // for json
	for t := range stream {
		if t.Err != nil {
			logrus.Infof("error in stream: %v", t.Err)
			http.Error(w, "error in stream", http.StatusInternalServerError)
			return
		}
		logrus.Debugf("token: %s", t.Token)
		if err := i.Put(t.Token); err != nil {
			logrus.Infof("could not indet token: %v", err)
			http.Error(w, "could not index", http.StatusInternalServerError)
			return
		}
		info.Tokens = append(info.Tokens, TokenInfo{
			Token:      t.Token.Token,
			ConceptURL: t.Token.Concept.URL(),
			Begin:      t.Token.Begin,
			End:        t.Token.End,
		})
	}
	e := json.NewEncoder(w)
	if err := e.Encode(&info); err != nil {
		logrus.Infof("could not encode json: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
	}
	logrus.Infof("handled %s", r.URL.Path)
}

func makeStream(dfa semix.DFA, r io.Reader) (semix.Stream, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	stream := semix.Filter(ctx,
		semix.Match(ctx, semix.DFAMatcher{DFA: dfa},
			semix.Read(ctx, r)))
	return stream, cancel
}
