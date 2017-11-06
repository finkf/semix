package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/query"
	"bitbucket.org/fflo/semix/pkg/semix"
)

type handle struct {
	g   *semix.Graph
	d   semix.Dictionary
	i   index.Index
	dfa semix.DFA
}

func newHandle(parser semix.Parser, traits semix.Traits, index index.Index) (handle, error) {
	log.Printf("reading knowledge base")
	graph, dictionary, err := semix.Parse(parser, traits)
	if err != nil {
		return handle{}, err
	}
	dfa := semix.NewDFA(dictionary, graph)
	log.Printf("done reading knowledge base")
	return handle{dfa: dfa, g: graph, d: dictionary, i: index}, nil
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
		log.Printf("serving request for %s", r.RequestURI)
		f(w, r)
		log.Printf("served request for %s", r.RequestURI)
	}
}

func (h handle) search(r *http.Request) (interface{}, int, error) {
	if r.Method != http.MethodGet {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid request method: %s", r.Method)
	}
	q := r.URL.Query().Get("q")
	if len(q) == 0 {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query: %v", q)
	}
	cs := Search(h.g, h.d, q)
	return cs, http.StatusOK, nil
}

func (h handle) info(r *http.Request) (interface{}, int, error) {
	if r.Method != http.MethodGet {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid request method: %s", r.Method)
	}
	qs := r.URL.Query()["q"]
	if len(qs) != 1 {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query: %v", qs)
	}
	q, err := url.QueryUnescape(qs[0])
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid url: %s", qs[0])
	}
	c, found := h.g.FindByURL(q)
	if !found {
		return nil, http.StatusNotFound, fmt.Errorf("invalid url: %s", q)
	}
	entries := SearchDictionaryEntries(h.d, c)
	info := ConceptInfo{Concept: c, Entries: entries}
	return info, http.StatusOK, nil
}

func (h handle) put(r *http.Request) (interface{}, int, error) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid request method: %s", r.Method)
	}
	doc, err := makeDocument(r)
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
				fmt.Errorf("cannot index document: %v", err)
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
	if len(r.URL.Query()["q"]) != 1 {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query parameter q=%v", r.URL.Query()["q"])
	}
	q, err := query.NewFix(r.URL.Query()["q"][0], func(arg string) (string, error) {
		cs := Search(h.g, h.d, arg)
		if len(cs) == 0 {
			return "", fmt.Errorf("cannot find %q", arg)
		}
		return cs[0].URL(), nil
	})
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid query: %v", err)
	}
	es, err := q.Execute(h.i)
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not execute query %q: %v", q, err)
	}
	var ts Tokens
	for _, e := range es {
		t, err := NewTokenFromEntry(h.g, e)
		if err != nil {
			return nil, http.StatusInternalServerError,
				fmt.Errorf("cannot convert %v: %v", e, err)
		}
		ts.Tokens = append(ts.Tokens, t)
	}
	return ts, http.StatusOK, nil
}

func (h handle) ctx(r *http.Request) (interface{}, int, error) {
	if r.Method != http.MethodGet {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid method: %v", r.Method)
	}
	url := r.URL.Query().Get("url")
	b, err1 := strconv.ParseInt(r.URL.Query().Get("b"), 10, 32)
	e, err2 := strconv.ParseInt(r.URL.Query().Get("e"), 10, 32)
	n, err3 := strconv.ParseInt(r.URL.Query().Get("n"), 10, 32)
	if url == "" || err1 != nil || err2 != nil || err3 != nil ||
		b < 0 || e < 0 || n < 0 || b > e {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query parameters = %s %v %v",
				url, []error{err1, err2, err3}, []int64{b, e, n})
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := semix.Normalize(ctx, semix.Read(ctx, semix.NewHTTPDocument(url)))
	t := <-s
	if t.Err != nil {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid document %s: %v", url, t.Err)
	}
	if int(b) >= len(t.Token.Token) || int(e) >= len(t.Token.Token) {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query paramters = %d %d", b, e)
	}
	cs := b - n
	if cs < 0 {
		cs = 0
	}
	ce := e + n
	if int(ce) > len(t.Token.Token) {
		ce = int64(len(t.Token.Token))
	}
	return Context{
		URL:    url,
		Before: t.Token.Token[cs:b],
		Match:  t.Token.Token[b:e],
		After:  t.Token.Token[e:ce],
		Begin:  int(b),
		End:    int(e),
		Len:    int(n),
	}, http.StatusOK, nil
}

func (h handle) makeIndexStream(d semix.Document) (semix.Stream, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	s := index.Put(ctx, h.i,
		semix.Filter(ctx,
			// semix.Match(ctx, semix.FuzzyDFAMatcher{DFA: semix.NewFuzzyDFA(3, h.dfa)},
			semix.Match(ctx, semix.DFAMatcher{DFA: h.dfa},
				semix.Normalize(ctx,
					semix.Read(ctx, d))))) // )
	return s, cancel
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
