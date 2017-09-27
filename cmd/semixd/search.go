package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"bitbucket.org/fflo/semix/pkg/semix"
	"github.com/sirupsen/logrus"
)

// SearchInfo represents the the results of a search.
type SearchInfo struct {
	Query   string
	Subject string
	Links   map[string][]string
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
	c := lookup(g, d, q[0])
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

func makeLookupInfo(d map[string]*semix.Concept, c *semix.Concept) SearchInfo {
	info := SearchInfo{Subject: c.URL()}
	for str, cc := range d {
		if cc == c {
			info.Entries = append(info.Entries, str)
		}
	}
	info.Links = make(map[string][]string)
	c.Edges(func(edge semix.Edge) {
		info.Links[edge.P.URL()] = append(info.Links[edge.P.URL()], edge.O.URL())
	})
	return info
}

func lookup(g *semix.Graph, d map[string]*semix.Concept, q string) *semix.Concept {
	if c, ok := g.FindByURL(q); ok {
		return c
	}
	if c, ok := d[q]; ok {
		return c
	}
	for e, c := range d {
		if strings.Contains(c.URL(), q) {
			return c
		}
		if strings.Contains(e, q) {
			return c
		}
	}
	return nil
}
