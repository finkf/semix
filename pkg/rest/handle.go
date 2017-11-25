package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/query"
	"bitbucket.org/fflo/semix/pkg/searcher"
	"bitbucket.org/fflo/semix/pkg/semix"
)

type lookupData struct {
	URL string
	ID  int
}

type handle struct {
	searcher  searcher.Searcher
	index     index.Interface
	dfa       semix.DFA
	dir, host string
}

func requestFunc(h func(*http.Request) (interface{}, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, status, err := h(r)
		if err != nil {
			writeError(w, status, err)
			return
		}
		buffer := new(bytes.Buffer)
		if err := json.NewEncoder(buffer).Encode(data); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf("could not encode: %v", err))
			return
		}
		w.WriteHeader(status)
		w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
		if _, err := w.Write(buffer.Bytes()); err != nil {
			log.Printf("could not write response: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	log.Printf("error: %v [%d]", err, status)
	w.Header()["Content-Type"] = []string{"text/plain; charset=utf-8"}
	http.Error(w, err.Error(), status)
}

func withLogging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("serving request for [%s] %s", r.Method, r.RequestURI)
		f(w, r)
		log.Printf("served request for [%s] %s", r.Method, r.RequestURI)
	}
}

func withGet(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusBadRequest,
				fmt.Errorf("invalid request method: %s", r.Method))
			return
		}
		f(w, r)
	}
}

func (h handle) parents(r *http.Request) (interface{}, int, error) {
	var data lookupData
	if err := DecodeQuery(r.URL.Query(), &data); err != nil {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query: %s", err)
	}
	c, ok := h.lookup(data)
	if !ok {
		return []*semix.Concept{}, http.StatusOK, nil
	}
	cs := h.searcher.SearchParents(c, -1)
	return cs, http.StatusOK, nil
}

func (h handle) search(r *http.Request) (interface{}, int, error) {
	q := r.URL.Query().Get("q")
	if len(q) == 0 {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query: %v", q)
	}
	cs := h.searcher.SearchConcepts(q, -1)
	return cs, http.StatusOK, nil
}

func (h handle) info(r *http.Request) (interface{}, int, error) {
	if r.Method != http.MethodGet {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid request method: %s", r.Method)
	}
	var data lookupData
	if err := DecodeQuery(r.URL.Query(), &data); err != nil {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query: %s", err)
	}
	c, ok := h.lookup(data)
	if !ok {
		return ConceptInfo{}, http.StatusOK, nil
	}
	entries := h.searcher.SearchDictionaryEntries(c)
	info := ConceptInfo{Concept: c, Entries: entries}
	return info, http.StatusOK, nil
}

func (h handle) put(r *http.Request) (interface{}, int, error) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid request method: %s", r.Method)
	}
	doc, err := h.makeDocument(r)
	if err != nil {
		return nil, http.StatusBadRequest,
			fmt.Errorf("bad document: %v", err)
	}
	stream, cancel := h.makeIndexStream(doc)
	defer cancel()
	ts := Tokens{Tokens: []Token{}} // for json
	for t := range stream {
		if t.Err != nil {
			return nil, http.StatusInternalServerError,
				fmt.Errorf("cannot index document: %v", t.Err)
		}
		ts.Tokens = append(ts.Tokens, NewTokens(t.Token)...)
	}
	return ts, http.StatusCreated, nil
}

func (h handle) get(r *http.Request) (interface{}, int, error) {
	if r.Method != http.MethodGet {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid method: %s", r.Method)
	}
	q := r.URL.Query().Get("q")
	log.Printf("query: %s", q)
	qu, err := query.New(q, h.getFixFunc())
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid query: %v", err)
	}
	log.Printf("executing query: %s", qu)
	es, err := qu.Execute(h.index)
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not execute query %q: %v", q, err)
	}
	var ts Tokens
	for _, e := range es {
		t, err := NewTokenFromEntry(h.searcher, e)
		if err != nil {
			return nil, http.StatusInternalServerError,
				fmt.Errorf("cannot convert %v: %v", e, err)
		}
		ts.Tokens = append(ts.Tokens, t)
	}
	return ts, http.StatusOK, nil
}

func (h handle) getFixFunc() query.LookupFunc {
	return func(arg string) ([]string, error) {
		cs := h.searcher.SearchConcepts(arg, 1)
		if len(cs) == 0 {
			return nil, fmt.Errorf("cannot find %q", arg)
		}
		var urls []string
		for _, c := range cs {
			if c.Ambiguous() {
				for i := 0; i < c.EdgesLen(); i++ {
					e := c.EdgeAt(i)
					urls = append(urls, e.O.URL())
				}
			} else {
				urls = append(urls, c.URL())
			}
		}
		return urls, nil
	}
}

func (h handle) ctx(r *http.Request) (interface{}, int, error) {
	if r.Method != http.MethodGet {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid method: %v", r.Method)
	}
	var data struct {
		URL     string
		B, E, N int
	}
	if err := DecodeQuery(r.URL.Query(), data); err != nil {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query parameters: %s", err)
	}
	t, err := h.readToken(data.URL)
	if err != nil {
		return nil, http.StatusNotFound,
			fmt.Errorf("invalid document %s: %v", data.URL, err)
	}
	if data.B >= len(t.Token) || data.E >= len(t.Token) {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query paramters = %d %d", data.B, data.E)
	}
	cs := data.B - data.N
	if cs < 0 {
		cs = 0
	}
	ce := data.E + data.N
	if int(ce) > len(t.Token) {
		ce = len(t.Token)
	}
	return Context{
		URL:    data.URL,
		Before: t.Token[cs:data.B],
		Match:  t.Token[data.B:data.E],
		After:  t.Token[data.E:ce],
		Begin:  int(data.B),
		End:    int(data.E),
		Len:    int(data.N),
	}, http.StatusOK, nil
}

func (h handle) readToken(url string) (semix.Token, error) {
	var d semix.Document
	if strings.HasPrefix(url, "semix-") {
		d = openDumpFile(h.dir, url)
	} else {
		d = semix.NewHTTPDocument(url)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := semix.Normalize(ctx, semix.Read(ctx, d))
	t := <-s
	if t.Err != nil {
		return semix.Token{}, t.Err
	}
	return t.Token, nil
}

func (h handle) makeIndexStream(d semix.Document) (semix.Stream, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	s := index.Put(ctx, h.index,
		semix.Filter(ctx,
			// semix.Match(ctx, semix.FuzzyDFAMatcher{DFA: semix.NewFuzzyDFA(3, h.dfa)},
			semix.Match(ctx, semix.DFAMatcher{DFA: h.dfa},
				semix.Normalize(ctx,
					semix.Read(ctx, d))))) // )
	return s, cancel
}

func (h handle) makeDocument(r *http.Request) (semix.Document, error) {
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
		r := strings.NewReader(strings.Join(r.PostForm["text"], " "))
		doc, err := newDumpFile(r, h.dir, "text/plain")
		if err != nil {
			return nil, fmt.Errorf("could not create file: %v", err)
		}
		return doc, nil
	}
}

func (h handle) lookup(data lookupData) (*semix.Concept, bool) {
	if data.ID != 0 {
		if c, ok := h.searcher.FindByID(data.ID); ok {
			return c, true
		}
	}
	if data.URL != "" {
		if c, ok := h.searcher.FindByURL(data.URL); ok {
			return c, true
		}
	}
	return nil, false
}
