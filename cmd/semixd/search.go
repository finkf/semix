package main

import (
	"encoding/json"
	"log"
	"net/http"

	"bitbucket.org/fflo/semix/pkg/net"
	"bitbucket.org/fflo/semix/pkg/semix"
)

func search(g *semix.Graph, d map[string]*semix.Concept, w http.ResponseWriter, r *http.Request) {
	log.Printf("serving request for %s", r.RequestURI)
	if r.Method != "GET" {
		log.Printf("invalid method: %s", r.Method)
		http.Error(w, "not a GET request", http.StatusBadRequest)
		return
	}
	q := r.URL.Query()["q"]
	if len(q) != 1 {
		log.Printf("invalid query: %v", q)
		http.Error(w, "invalid query paramters", http.StatusBadRequest)
		return
	}
	// if c cannot be found; it is nil. SearchDictionaryEntries handles this case.
	c, _ := net.Search(g, d, q[0])
	entries := net.SearchDictionaryEntries(d, c)
	info := net.ConceptInfo{Concept: c, Entries: entries}
	if err := json.NewEncoder(w).Encode(info); err != nil {
		log.Printf("could not encode json: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	log.Printf("handled %s", r.URL.Path)
}
