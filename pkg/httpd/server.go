package httpd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"bitbucket.org/fflo/semix/pkg/client"
	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rest"
	"bitbucket.org/fflo/semix/pkg/say"
	x "bitbucket.org/fflo/semix/pkg/semix"
	"github.com/pkg/errors"
)

type status struct {
	err    error
	status int
}

type httpdHandle func(*http.Request) (*template.Template, interface{}, status)

func handle(f httpdHandle) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, x, s := f(r)
		if s.err != nil {
			say.Info("could not handle request: %v", s.err)
			http.Error(w, s.err.Error(), s.status)
			return
		}
		buffer := new(bytes.Buffer)
		if err := t.Execute(buffer, x); err != nil {
			say.Info("could not execute template: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(s.status)
		w.Header()["Content-Type"] = []string{"text/html; charset=utf-8"}
		if _, err := w.Write(buffer.Bytes()); err != nil {
			say.Info("could not write html: %v", err)
		}
	}
}

func (s *Server) favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.dir, "favicon.ico"))
}

func (s *Server) semixJS(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.dir, "js", "semix.js"))
}

func (s *Server) semixCSS(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.dir, "css", "semix.css"))
}

func (s *Server) httpdSearch(r *http.Request) (*template.Template, interface{}, status) {
	q := r.URL.Query().Get("q")
	cs, err := s.newClient().Search(q)
	if err != nil {
		return nil, nil, internalError(err)
	}
	return s.searchtmpl, struct {
		Title    string
		Concepts []*x.Concept
	}{fmt.Sprintf("%q", q), cs}, ok()
}

func (s *Server) predicates(r *http.Request) (*template.Template, interface{}, status) {
	q := r.URL.Query().Get("q")
	cs, err := s.newClient().Predicates(q)
	if err != nil {
		return nil, nil, internalError(err)
	}
	return s.searchtmpl, struct {
		Title    string
		Concepts []*x.Concept
	}{fmt.Sprintf("%q", q), cs}, ok()
}

func (s *Server) parents(r *http.Request) (*template.Template, interface{}, status) {
	q := r.URL.Query().Get("url")
	cs, err := s.newClient().ParentsURL(q)
	if err != nil {
		return nil, nil, internalError(err)
	}
	return s.searchtmpl, struct {
		Title    string
		Concepts []*x.Concept
	}{fmt.Sprintf("parents of %q", q), cs}, ok()
}

func (s *Server) httpdInfo(r *http.Request) (*template.Template, interface{}, status) {
	q := r.URL.Query().Get("url")
	info, err := s.newClient().InfoURL(q)
	if err != nil {
		return nil, nil, internalError(err)
	}
	return s.infotmpl, info, ok()
}

func (s *Server) home(r *http.Request) (*template.Template, interface{}, status) {
	if strings.HasPrefix(r.URL.RequestURI(), "/semix-") {
		url, err := url.QueryUnescape(r.URL.RequestURI())
		if err != nil {
			return nil, nil, internalError(err)
		}
		say.Info("r.URL.RequestURI(): %q", url)
		content, err := s.newClient().DumpFile(url)
		if err != nil {
			return nil, nil, internalError(err)
		}
		return s.dumptmpl, content, ok()
	}
	return s.indextmpl, nil, ok()
}

func (s *Server) httpdGet(r *http.Request) (*template.Template, interface{}, status) {
	var data struct {
		Q    string
		N, S int
	}
	if err := rest.DecodeQuery(r.URL.Query(), &data); err != nil {
		return nil, nil, internalError(err)
	}
	es, err := s.newClient(client.WithMax(data.N), client.WithSkip(data.S)).Get(data.Q)
	if err != nil {
		return nil, nil, internalError(err)
	}
	return s.gettmpl, struct {
		Query   string
		Entries []index.Entry
		N, S    int
	}{data.Q, es, data.N, data.S}, ok()
}

func (s *Server) ctx(r *http.Request) (*template.Template, interface{}, status) {
	var ctx rest.Context
	var data struct {
		URL     string
		B, E, N int
	}
	if err := rest.DecodeQuery(r.URL.Query(), &data); err != nil {
		return nil, nil, internalError(err)
	}
	ctx, err := s.newClient().Ctx(data.URL, data.B, data.E, data.N)
	if err != nil {
		return nil, nil, internalError(err)
	}
	return s.ctxtmpl, ctx, ok()
}

func (s *Server) httpdPut(r *http.Request) (*template.Template, interface{}, status) {
	var data rest.PutData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return nil, nil, internalError(errors.Wrapf(err, "could not decode post data"))
	}
	say.Debug("got request for %v", data)
	ts, err := s.newClient(
		client.WithErrorLimits(data.Errors...),
		client.WithResolvers(data.Resolvers...),
	).PutContent(strings.NewReader(data.Content), data.URL, data.ContentType)
	if err != nil {
		return nil, nil, internalError(errors.Wrapf(err, "could not put content"))
	}
	return s.puttmpl, ts, ok()
}

func (s *Server) setup(r *http.Request) (*template.Template, interface{}, status) {
	return s.setuptmpl, struct{}{}, ok()
}

func internalError(err error) status {
	return status{err, http.StatusInternalServerError}
}

func ok() status {
	return status{nil, http.StatusOK}
}
