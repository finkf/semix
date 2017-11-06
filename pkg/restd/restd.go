package restd

import (
	"net/http"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/semix"
)

// New returns a new server instance.
func New(self string, g *semix.Graph, d semix.Dictionary, i index.Interface) *http.Server {
	return &http.Server{
		Addr:    self,
		Handler: newMux(g, d, i),
	}
}

func newMux(g *semix.Graph, d semix.Dictionary, i index.Interface) *http.ServeMux {
	dfa := semix.NewDFA(d, g)
	h := handle{dfa: dfa, d: d, g: g, i: i}
	mux := http.NewServeMux()
	mux.HandleFunc("/search", withLogging(requestFunc(h.search)))
	mux.HandleFunc("/put", withLogging(requestFunc(h.put)))
	mux.HandleFunc("/get", withLogging(requestFunc(h.get)))
	mux.HandleFunc("/ctx", withLogging(requestFunc(h.ctx)))
	mux.HandleFunc("/info", withLogging(requestFunc(h.info)))
	return mux
}
