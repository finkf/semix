package rest

import (
	"net/http"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/semix"
)

// New returns a new server instance.
func New(self string, parser semix.Parser, traits semix.Traits, index index.Index) (*http.Server, error) {
	mux, err := newMux(parser, traits, index)
	if err != nil {
		return nil, err
	}
	return &http.Server{
		Addr:    self,
		Handler: mux,
	}, nil
}

func newMux(parser semix.Parser, traits semix.Traits, index index.Index) (*http.ServeMux, error) {
	h, err := newHandle(parser, traits, index)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/search", withLogging(requestFunc(h.search)))
	mux.HandleFunc("/put", withLogging(requestFunc(h.put)))
	mux.HandleFunc("/get", withLogging(requestFunc(h.get)))
	mux.HandleFunc("/ctx", withLogging(requestFunc(h.ctx)))
	mux.HandleFunc("/info", withLogging(requestFunc(h.info)))
	return mux, nil
}
