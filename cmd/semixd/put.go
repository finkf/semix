package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/net"
	"bitbucket.org/fflo/semix/pkg/semix"
)

func put(dfa semix.DFA, i index.Index, w http.ResponseWriter, r *http.Request) {
	log.Printf("serving request for %s", r.RequestURI)
	if r.Method != "POST" && r.Method != "GET" {
		log.Printf("invalid method: %s", r.Method)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	doc, err := makeDocumentFromRequest(r)
	if err != nil {
		log.Printf("could not create document: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	stream, cancel := makeStream(dfa, doc)
	defer cancel()
	ts := net.Tokens{Tokens: []semix.Token{}} // for json
	for t := range stream {
		if t.Err != nil {
			log.Printf("error in stream: %v", t.Err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		// logrus.Debugf("token: %s", t.Token)
		if err := i.Put(t.Token); err != nil {
			log.Printf("could not index token: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		ts.Tokens = append(ts.Tokens, t.Token)
	}
	e := json.NewEncoder(w)
	if err := e.Encode(&ts); err != nil {
		log.Printf("could not encode json: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
	}
	log.Printf("handled %s", r.URL.Path)
}

func makeStream(dfa semix.DFA, d semix.Document) (semix.Stream, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	stream := semix.Filter(ctx,
		semix.Match(ctx, semix.DFAMatcher{DFA: dfa},
			semix.Read(ctx, d)))
	return stream, cancel
}

func makeDocumentFromRequest(r *http.Request) (semix.Document, error) {
	if r.Method == "GET" {
		url := r.URL.Query()["url"]
		if len(url) != 1 {
			return nil, errors.New("missing url query paramter")
		}
		return semix.NewHTTPDocument(url[0]), nil
	}
	path := time.Now().Format(time.RFC3339) + "-" + r.RemoteAddr
	doc, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	repl := strings.NewReplacer("=", " ", "+", " ")
	return requestDocument{
		r:    strings.NewReader(repl.Replace(string(doc))),
		path: path,
	}, nil
}

type requestDocument struct {
	r    io.Reader
	path string
}

func (r requestDocument) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r requestDocument) Close() error {
	return nil
}

func (r requestDocument) Path() string {
	return r.path
}
