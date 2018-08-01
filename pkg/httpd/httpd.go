// Package httpd implements a simple http server for semix
package httpd

import (
	"context"
	"html/template"
	"net/http"
	"strings"

	"gitlab.com/finkf/semix/pkg/client"
	"gitlab.com/finkf/semix/pkg/rest"
)

// Option is a functional option to configure the server
type Option func(*Server)

// WithDirectory set the base directory for the server files.
func WithDirectory(dir string) Option {
	return func(s *Server) {
		s.dir = dir
	}
}

// WithHost sets the host address the server listens on.
func WithHost(host string) Option {
	return func(s *Server) {
		s.host = host
	}
}

// WithDaemon sets the address of the semix rest daemon.
func WithDaemon(daemon string) Option {
	return func(s *Server) {
		host := daemon
		if !strings.HasPrefix(host, "http://") ||
			!strings.HasPrefix(host, "https://") {
			host = "http://" + host
		}
		s.daemon = host
	}
}

// Server is the httpd server.
type Server struct {
	host, dir  string
	daemon     string
	server     *http.Server
	infotmpl   *template.Template
	puttmpl    *template.Template
	indextmpl  *template.Template
	ctxtmpl    *template.Template
	searchtmpl *template.Template
	dumptmpl   *template.Template
	gettmpl    *template.Template
	setuptmpl  *template.Template
}

// New returns a new server with a default configuration.
// Use options to configure the server.
func New(opts ...Option) (*Server, error) {
	s := &Server{
		host: "localhost:80",
		dir:  "html",
	}
	for _, opt := range opts {
		opt(s)
	}
	s.server = newMux(s)
	err := new(setuper).setup(s)
	return s, err
}

// Start starts the server.
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Close closes the server and its enclosed index.
func (s *Server) Close() error {
	return s.server.Shutdown(context.TODO())
}

func newMux(s *Server) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/",
		rest.WithLogging(rest.WithGet(handle(s.home))))
	mux.HandleFunc("/index",
		rest.WithLogging(rest.WithGet(handle(s.home))))
	mux.HandleFunc("/info",
		rest.WithLogging(rest.WithGet(handle(s.info))))
	mux.HandleFunc("/get",
		rest.WithLogging(rest.WithGet(handle(s.get))))
	mux.HandleFunc("/search",
		rest.WithLogging(rest.WithGet(handle(s.search))))
	mux.HandleFunc("/predicates",
		rest.WithLogging(rest.WithGet(handle(s.predicates))))
	mux.HandleFunc("/ctx",
		rest.WithLogging(rest.WithGet(handle(s.ctx))))
	mux.HandleFunc("/put",
		rest.WithLogging(rest.WithGetOrPost(handle(s.put))))
	mux.HandleFunc("/setup",
		rest.WithLogging(rest.WithGet(handle(s.setup))))
	mux.HandleFunc("/parents",
		rest.WithLogging(rest.WithGet(handle(s.parents))))
	// do not log requests for favicon
	mux.HandleFunc("/favicon.ico", rest.WithGet(s.favicon))
	mux.HandleFunc("/js/semix.js",
		rest.WithLogging(rest.WithGet(s.js)))
	mux.HandleFunc("/graph.svg", rest.WithLogging(rest.WithGet(s.graph)))
	return &http.Server{
		Addr:    s.host,
		Handler: mux,
	}
}

func (s *Server) newClient(opts ...client.Option) *client.Client {
	return client.New(s.daemon, opts...)
}
