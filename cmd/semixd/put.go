package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/semix"
	"github.com/sirupsen/logrus"
)

type TokenInfo struct {
	Token, ConceptURL, Path string
	Begin, End              int
	Links                   map[string][]string
}

type IndexInfo struct {
	Tokens []TokenInfo
}

func put(dfa semix.DFA, i index.Index, w http.ResponseWriter, r *http.Request) {
	logrus.Infof("serving request for %s", r.RequestURI)
	if r.Method != "POST" && r.Method != "GET" {
		logrus.Infof("invalid method: %s", r.Method)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	doc, err := makeDocumentFromRequest(r)
	if err != nil {
		logrus.Infof("could not create document: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	stream, cancel := makeStream(dfa, doc)
	defer cancel()
	info := IndexInfo{Tokens: []TokenInfo{}} // for json
	for t := range stream {
		if t.Err != nil {
			logrus.Infof("error in stream: %v", t.Err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		// logrus.Debugf("token: %s", t.Token)
		if err := i.Put(t.Token); err != nil {
			logrus.Infof("could not index token: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		token := TokenInfo{
			Token:      t.Token.Token,
			ConceptURL: t.Token.Concept.URL(),
			Path:       t.Token.Path,
			Begin:      t.Token.Begin,
			End:        t.Token.End,
			Links:      make(map[string][]string),
		}
		for i := 0; i < t.Token.Concept.EdgesLen(); i++ {
			edge := t.Token.Concept.EdgeAt(i)
			token.Links[edge.P.URL()] = append(token.Links[edge.P.URL()], edge.O.URL())
		}
		info.Tokens = append(info.Tokens, token)
	}
	e := json.NewEncoder(w)
	if err := e.Encode(&info); err != nil {
		logrus.Infof("could not encode json: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
	}
	logrus.Infof("handled %s", r.URL.Path)
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
