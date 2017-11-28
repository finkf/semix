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
	"bitbucket.org/fflo/semix/pkg/resolve"
	"bitbucket.org/fflo/semix/pkg/searcher"
	"bitbucket.org/fflo/semix/pkg/semix"
)

type lookupData struct {
	URL string
	ID  int
}

type putData struct {
	URL string
	T   float64
	N   int
	L   []int
	R   []string
}

type handle struct {
	searcher  searcher.Searcher
	index     index.Interface
	dir, host string
	r         *semix.Resource
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

func (h handle) predicates(r *http.Request) (interface{}, int, error) {
	q := r.URL.Query().Get("q")
	if len(q) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid query: %s", q)
	}
	cs := h.searcher.SearchPredicates(q, -1)
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
	var data putData
	if err := DecodeQuery(r.URL.Query(), &data); err != nil {
		return nil, http.StatusBadRequest, err
	}
	doc, err := h.makeDocument(data, r.Method == http.MethodGet)
	if err != nil {
		return nil, http.StatusBadRequest,
			fmt.Errorf("bad document: %v", err)
	}
	stream, cancel, err := h.indexer(data, doc)
	defer cancel()
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
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

func (h handle) indexer(data putData, doc semix.Document) (semix.Stream, context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s := h.matcher(ctx, data, semix.Normalize(ctx, semix.Read(ctx, doc)))
	s, err := h.resolver(ctx, data, s)
	if err != nil {
		return nil, cancel, err
	}
	return index.Put(ctx, h.index, semix.Filter(ctx, s)), cancel, nil
}

func (h handle) resolver(ctx context.Context, d putData, s semix.Stream) (semix.Stream, error) {
	for i := len(d.R); i > 0; i-- {
		r, err := h.resolverInterface(d.R[i-1], d)
		if err != nil {
			return nil, err
		}
		s = resolve.Resolve(ctx, d.N, r, s)
	}
	return s, nil
}

func (h handle) resolverInterface(name string, d putData) (resolve.Interface, error) {
	switch name {
	case "automatic":
		return resolve.Automatic{Threshold: d.T}, nil
	case "simple":
		return resolve.Simple{}, nil
	case "ruled":
		return resolve.NewRuled(h.r.Rules, func(str string) int {
			cs := h.searcher.SearchConcepts(str, 2)
			if len(cs) != 1 {
				return -1
			}
			return int(cs[0].ID())
		})
	}
	return nil, fmt.Errorf("invalid resolver: %s", name)
}

func (h handle) matcher(ctx context.Context, d putData, s semix.Stream) semix.Stream {
	for i := len(d.L); i > 0; i-- {
		l := d.L[i-1]
		if l <= 0 {
			continue
		}
		s = semix.Match(ctx, semix.FuzzyDFAMatcher{DFA: semix.NewFuzzyDFA(l, h.r.DFA)}, s)
	}
	return semix.Match(ctx, semix.DFAMatcher{DFA: h.r.DFA}, s)
}

func (h handle) makeDocument(data putData, get bool) (semix.Document, error) {
	if get {
		if len(data.URL) <= 0 {
			return nil, fmt.Errorf("invalid query parameter url=%s", data.URL)
		}
		return semix.NewHTTPDocument(data.URL), nil
	}
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
