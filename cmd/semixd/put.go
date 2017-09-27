package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/semix"
	"github.com/sirupsen/logrus"
)

type TokenInfo struct {
	Token, ConceptURL, Path string
	Begin, End              int
}

type IndexInfo struct {
	Tokens []TokenInfo
}

func put(dfa semix.DFA, i index.Index, w http.ResponseWriter, r *http.Request) {
	logrus.Infof("serving request for %s", r.RequestURI)
	if r.Method != "POST" {
		logrus.Infof("invalid method: %s", r.Method)
		http.Error(w, "not a POST request", http.StatusBadRequest)
		return
	}
	stream, cancel := makeStream(dfa, makeDocumentFromRequest(r))
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
			Path:       t.Token.Path,
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

func makeStream(dfa semix.DFA, d semix.Document) (semix.Stream, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	stream := semix.Filter(ctx,
		semix.Match(ctx, semix.DFAMatcher{DFA: dfa},
			semix.Read(ctx, d)))
	return stream, cancel
}

func makeDocumentFromRequest(r *http.Request) semix.Document {
	path := time.Now().Format(time.RFC3339) + "-" + r.RemoteAddr
	return requestDocument{
		r:    r.Body,
		path: path,
	}
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
