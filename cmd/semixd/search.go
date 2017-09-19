package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"bitbucket.org/fflo/semix/pkg/semix"
	"github.com/sirupsen/logrus"
)

type LinkInfo struct {
	Predicate, Object string
}

type SearchInfo struct {
	Query   string
	Subject string
	Links   []LinkInfo
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

func makeLookupInfo(d map[string]*semix.Concept, c *semix.Concept) SearchInfo {
	info := SearchInfo{Subject: c.URL()}
	for str, cc := range d {
		if cc == c {
			info.Entries = append(info.Entries, str)
		}
	}
	c.Edges(func(edge semix.Edge) {
		info.Links = append(info.Links, LinkInfo{
			Predicate: edge.P.URL(),
			Object:    edge.O.URL(),
		})
	})
	return info
}

func lookup(d map[string]*semix.Concept, q string) *semix.Concept {
	if c, ok := d[q]; ok {
		return c
	}
	for _, c := range d {
		if strings.Contains(c.URL(), q) {
			return c
		}
	}
	return nil
}
