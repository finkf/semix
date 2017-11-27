package rest

import (
	"context"
	"net/http"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/searcher"
	"bitbucket.org/fflo/semix/pkg/semix"
)

// Server represents a server instance.
type Server struct {
	server *http.Server
	handle handle
}

// New returns a new server instance.
func New(self, dir string, r *semix.Resource, i index.Interface) *Server {
	dfa := semix.NewDFA(r.Dictionary, r.Graph)
	h := handle{
		dir:      dir,
		dfa:      dfa,
		r:        r,
		searcher: searcher.New(r.Graph, r.Dictionary),
		index:    i,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/search", withLogging(withGet(requestFunc(h.search))))
	mux.HandleFunc("/parents", withLogging(withGet(requestFunc(h.parents))))
	mux.HandleFunc("/put", withLogging(requestFunc(h.put)))
	mux.HandleFunc("/get", withLogging(requestFunc(h.get)))
	mux.HandleFunc("/ctx", withLogging(requestFunc(h.ctx)))
	mux.HandleFunc("/info", withLogging(requestFunc(h.info)))
	return &Server{
		server: &http.Server{
			Addr:    self,
			Handler: mux,
		},
		handle: h,
	}
}

// ListenAndServe starts the server.
func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// Close closes the server and its enclosed index.
func (s *Server) Close() error {
	err1 := s.handle.index.Close()
	err2 := s.server.Shutdown(context.TODO())
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}
