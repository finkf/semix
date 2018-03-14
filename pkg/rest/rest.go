package rest

import (
	"context"
	"net/http"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rule"
	"bitbucket.org/fflo/semix/pkg/searcher"
	"bitbucket.org/fflo/semix/pkg/semix"
	"github.com/pkg/errors"
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
	mux.HandleFunc("/concept", WithLogging(WithGet(requestFunc(h.concept))))
	mux.HandleFunc("/search", WithLogging(WithGet(requestFunc(h.search))))
	mux.HandleFunc("/parents", WithLogging(WithGet(requestFunc(h.parents))))
	mux.HandleFunc("/predicates", WithLogging(WithGet(requestFunc(h.predicates))))
	mux.HandleFunc("/put", WithLogging(WithPost(requestFunc(h.put))))
	mux.HandleFunc("/get", WithLogging(WithGet(requestFunc(h.get))))
	mux.HandleFunc("/ctx", WithLogging(WithGet(requestFunc(h.ctx))))
	mux.HandleFunc("/info", WithLogging(WithGet(requestFunc(h.info))))
	mux.HandleFunc("/dump", WithLogging(WithGet(requestFunc(h.dump))))
	mux.HandleFunc("/flush", WithLogging(WithGet(requestFunc(h.flush))))
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
	err := errors.Wrapf(s.handle.index.Close(), "cannot close index")
	err = errors.Wrapf(s.server.Shutdown(context.TODO()), "cannot shutdown server")
	return err
}
