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
	"bitbucket.org/fflo/semix/pkg/rule"
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
	dir, host string
	dfa       semix.DFA
	rules     rule.Map
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

func (h handle) concept(r *http.Request) (interface{}, int, error) {
	var data lookupData
	if err := DecodeQuery(r.URL.Query(), &data); err != nil {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query: %s", err)
	}
	c, ok := h.lookup(data)
	if !ok {
		return nil, http.StatusNotFound, nil
	}
	return c, http.StatusOK, nil
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
		return nil, http.StatusNotFound, nil
	}
	entries := h.searcher.SearchDictionaryEntries(c)
	info := ConceptInfo{Concept: c, Entries: entries}
	return info, http.StatusOK, nil
}

func (h handle) put(r *http.Request) (interface{}, int, error) {
	// do not check for content type, since the json decoding should
	// give an error.
	var data PutData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return nil, http.StatusBadRequest, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := data.stream(ctx, h.dfa, h.rules, h.index, h.dir)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	var es []index.Entry
	for t := range stream {
		if t.Err != nil {
			return nil, http.StatusInternalServerError,
				fmt.Errorf("cannot index document: %s", t.Err)
		}
		es = append(es, index.Entry{
			Path:       t.Token.Path,
			Token:      t.Token.Token,
			ConceptURL: t.Token.Concept.URL(),
			Begin:      t.Token.Begin,
			End:        t.Token.End,
		})
	}
	return es, http.StatusCreated, nil
}

func (h handle) get(r *http.Request) (interface{}, int, error) {
	var data struct {
		Q    string
		N, S int
	}
	if err := DecodeQuery(r.URL.Query(), &data); err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid query: %s", err)
	}
	q, err := query.New(data.Q, h.getFixFunc())
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid query: %s", err)
	}
	var es []index.Entry
	err = q.ExecuteFunc(h.index, func(e index.Entry) bool {
		if data.S > 0 {
			data.S--
			return true
		}
		if data.N == 0 || len(es) < data.N {
			es = append(es, e)
			return true
		}
		return false
	})
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not execute query %q: %v", q, err)
	}
	return es, http.StatusOK, nil
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
	if err := DecodeQuery(r.URL.Query(), &data); err != nil {
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
