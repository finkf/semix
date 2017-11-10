package restd

import (
	"net/http"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/searcher"
	"bitbucket.org/fflo/semix/pkg/semix"
)

// New returns a new server instance.
func New(self, dir string, g *semix.Graph, d semix.Dictionary, i index.Interface) *http.Server {
	return &http.Server{
		Addr:    self,
		Handler: newMux(dir, g, d, i),
	}
}

func newMux(dir string, g *semix.Graph, d semix.Dictionary, i index.Interface) *http.ServeMux {
	dfa := semix.NewDFA(d, g)
	h := handle{dir: dir, dfa: dfa, searcher: searcher.New(g, d), i: i}
	mux := http.NewServeMux()
	mux.HandleFunc("/search", withLogging(withGet(requestFunc(h.search))))
	mux.HandleFunc("/parents", withLogging(withGet(requestFunc(h.parents))))
	mux.HandleFunc("/put", withLogging(requestFunc(h.put)))
	mux.HandleFunc("/get", withLogging(requestFunc(h.get)))
	mux.HandleFunc("/ctx", withLogging(requestFunc(h.ctx)))
	mux.HandleFunc("/info", withLogging(requestFunc(h.info)))
	return mux
}
