package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/net"
	"bitbucket.org/fflo/semix/pkg/query"
	"bitbucket.org/fflo/semix/pkg/semix"
)

type handle struct {
	g   *semix.Graph
	d   map[string]*semix.Concept
	i   index.Index
	dfa semix.DFA
}

func requestFunc(h func(*http.Request) (interface{}, int, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, status, err := h(r)
		if err != nil {
			log.Printf("error: %v", err)
			w.Header()["Content-Type"] = []string{"text/plain", "charset=utf-8"}
			http.Error(w, err.Error(), status)
			return
		}
		buffer := new(bytes.Buffer)
		if err := json.NewEncoder(buffer).Encode(data); err != nil {
			log.Printf("could not encode: %v", err)
			http.Error(w, fmt.Sprintf("could not encode response: %v", err),
				http.StatusInternalServerError)
			return
		}
		w.WriteHeader(status)
		w.Header()["Content-Type"] = []string{"application/json", "charset=utf-8"}
		if _, err := w.Write(buffer.Bytes()); err != nil {
			log.Printf("could not write response: %v", err)
		}
	}
}

func (h handle) search(r *http.Request) (interface{}, int, error) {
	log.Printf("serving request for %s", r.RequestURI)
	if r.Method != http.MethodGet {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid request method: %s", r.Method)
	}
	q := r.URL.Query()["q"]
	if len(q) != 1 {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query: %v", q)
	}
	// if c cannot be found; it is nil.
	// SearchDictionaryEntries handles this case.
	c, _ := net.Search(h.g, h.d, q[0])
	entries := net.SearchDictionaryEntries(h.d, c)
	info := net.ConceptInfo{Concept: c, Entries: entries}
	log.Printf("handled %s", r.URL.Path)
	return info, http.StatusOK, nil
}

func (h handle) put(r *http.Request) (interface{}, int, error) {
	log.Printf("serving request for %s", r.RequestURI)
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid request method: %s", r.Method)
	}
	doc, err := makeDocument(r)
	if err != nil {
		return nil, http.StatusBadRequest,
			fmt.Errorf("bad document: %v", err)
	}
	stream := h.makeStream(doc)
	ts := net.Tokens{Tokens: []semix.Token{}} // for json
	for t := range stream {
		if t.Err != nil {
			return nil, http.StatusInternalServerError,
				fmt.Errorf("bad document: %v", err)
		}
		if err := h.i.Put(t.Token); err != nil {
			return nil, http.StatusInternalServerError,
				fmt.Errorf("cannot index token %q: %v", t.Token, err)
		}
		ts.Tokens = append(ts.Tokens, t.Token)
	}
	return ts, http.StatusCreated, nil
}

func (h handle) get(r *http.Request) (interface{}, int, error) {
	log.Printf("serving request for %s", r.RequestURI)
	if r.Method != http.MethodGet {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid method: %s", r.Method)
	}
	if len(r.URL.Query()["q"]) != 1 {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query parameter q=%v", r.URL.Query()["q"])
	}
	q, err := query.NewFix(r.URL.Query()["q"][0], func(arg string) (string, error) {
		c, ok := net.Search(h.g, h.d, arg)
		if !ok {
			return "", fmt.Errorf("cannot find %q", arg)
		}
		return c.URL(), nil
	})
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid query: %v", err)
	}
	es, err := q.Execute(h.i)
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not execute query %q: %v", q, err)
	}
	var ts net.Tokens
	for _, e := range es {
		t, err := net.NewTokenFromEntry(h.g, e)
		if err != nil {
			return nil, http.StatusInternalServerError,
				fmt.Errorf("cannot convert %v: %v", e, err)
		}
		ts.Tokens = append(ts.Tokens, t)
	}
	return ts, http.StatusOK, nil
}

func (h handle) makeStream(d semix.Document) semix.Stream {
	return semix.Filter(semix.Match(semix.DFAMatcher{DFA: h.dfa}, semix.Read(d)))
}

func makeDocument(r *http.Request) (semix.Document, error) {
	switch r.Method {
	default:
		panic("invalid method")
	case http.MethodGet:
		url := r.URL.Query()["url"]
		if len(url) != 1 {
			return nil, fmt.Errorf("invalid query parameter url=%v", url)
		}
		return semix.NewHTTPDocument(url[0]), nil
	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			return nil, fmt.Errorf("could not parse post form: %v", err)
		}
		path := time.Now().Format("2006-01-02:15-04-05")
		str := " " + strings.Join(r.PostForm["text"], " ") + " "
		str = regexp.MustCompile(`\s+`).ReplaceAllLiteralString(str, " ")
		return semix.NewStringDocument(path, str), nil
	}
}
