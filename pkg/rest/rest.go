package rest

import (
	"context"
	"net/http"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rule"
	"bitbucket.org/fflo/semix/pkg/searcher"
	"bitbucket.org/fflo/semix/pkg/semix"
)

// Server represents a server instance.
type Server struct {
	server *http.Server
	handle handle
}

// New returns a new server instance.
func New(self, dir string, r *semix.Resource, i index.Interface) (*Server, error) {
	searcher := searcher.New(r.Graph, r.Dictionary)
	rules, err := rule.NewMap(r.Rules, func(str string) int {
		cs := searcher.SearchConcepts(str, 2)
		if len(cs) != 1 {
			return -1
		}
		return int(cs[0].ID())
	})
	if err != nil {
		return nil, err
	}
	h := handle{
		dir:      dir,
		dfa:      r.DFA,
		searcher: searcher,
		rules:    rules,
		index:    i,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/search", WithLogging(WithGet(requestFunc(h.search))))
	mux.HandleFunc("/parents", WithLogging(WithGet(requestFunc(h.parents))))
	mux.HandleFunc("/predicates", WithLogging(WithGet(requestFunc(h.predicates))))
	mux.HandleFunc("/put", WithLogging(WithPost(requestFunc(h.put))))
	mux.HandleFunc("/get", WithLogging(WithGet(requestFunc(h.get))))
	mux.HandleFunc("/ctx", WithLogging(WithGet(requestFunc(h.ctx))))
	mux.HandleFunc("/info", WithLogging(WithGet(requestFunc(h.info))))
	return &Server{
		server: &http.Server{
			Addr:    self,
			Handler: mux,
		},
		handle: h,
	}, nil
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
