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

	"github.com/finkf/semix/pkg/client"
	"github.com/finkf/semix/pkg/dot"
	"github.com/finkf/semix/pkg/index"
	"github.com/finkf/semix/pkg/rest"
	"github.com/finkf/semix/pkg/say"
	"github.com/finkf/semix/pkg/semix"
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
			say.Info("cannot handle request: %v", s.err)
			http.Error(w, s.err.Error(), s.status)
			return
		}
		buffer := new(bytes.Buffer)
		if err := t.Execute(buffer, x); err != nil {
			say.Info("cannot execute template: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(s.status)
		w.Header()["Content-Type"] = []string{"text/html; charset=utf-8"}
		if _, err := w.Write(buffer.Bytes()); err != nil {
			say.Info("cannot write html: %v", err)
		}
	}
}

func (s *Server) favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.dir, "favicon.ico"))
}

func (s *Server) js(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.dir, "js", "semix.js"))
}

func (s *Server) css(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.dir, "css", "semix.css"))
}

func (s *Server) search(r *http.Request) (*template.Template, interface{}, status) {
	q := r.URL.Query().Get("q")
	cs, err := s.newClient().Search(q)
	if err != nil {
		return nil, nil, internalError(err)
	}
	return s.searchtmpl, struct {
		Title    string
		Concepts []*semix.Concept
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
		Concepts []*semix.Concept
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
		Concepts []*semix.Concept
	}{fmt.Sprintf("parents of %q", q), cs}, ok()
}

func (s *Server) info(r *http.Request) (*template.Template, interface{}, status) {
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

func (s *Server) get(r *http.Request) (*template.Template, interface{}, status) {
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

func (s *Server) put(r *http.Request) (*template.Template, interface{}, status) {
	var data rest.PutData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return nil, nil, internalError(errors.Wrapf(err, "cannot decode post data"))
	}
	if data.URL == "" && data.Content == "" {
		return nil, nil, badRequest(errors.New("URL and content missing"))
	}
	client := s.newClient(
		client.WithErrorLimits(data.Errors...),
		client.WithResolvers(data.Resolvers...),
	)
	var ts []index.Entry
	var err error
	if data.URL != "" {
		ts, err = client.PutURL(data.URL)
	} else {
		ts, err = client.PutContent(strings.NewReader(data.Content), data.URL, data.ContentType)
	}
	if err != nil {
		return nil, nil, internalError(errors.Wrapf(err, "cannot put content"))
	}
	return s.puttmpl, ts, ok()
}

func (s *Server) setup(r *http.Request) (*template.Template, interface{}, status) {
	return s.setuptmpl, struct{}{}, ok()
}

type link struct {
	S, P, O string
}

func (s *Server) graph(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	cs, err := s.newClient().DownloadURL(url)
	if err != nil {
		say.Info("cannot handle request: cannot download: %s: %v", url, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, ok := cs[url]; !ok {
		say.Info("cannot handle request: cannot find url: %s", url)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	cs[url].ReduceTransitive()
	d := dot.New(fmt.Sprintf("%q", cs[url].ShortName()), dot.Rankdir, dot.BT)
	cs[url].VisitAll(func(c *semix.Concept) {
		if c.URL() == url {
			d.AddNode(c.URL(), dot.Label, labelName(c), "shape", "box",
				dot.FillColor, "lightgrey", dot.Style, dot.Filled)
		} else {
			d.AddNode(c.URL(), dot.Label, labelName(c), "shape", "box")
		}
		for _, e := range c.Edges() {
			d.AddEdge(c.URL(), e.O.URL(), dot.Label, labelName(e.P))
		}
	})
	svg, err := d.SVG("/usr/bin/dot")
	if err != nil {
		say.Info("cannot handle request: cannot generate image: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	if _, err := w.Write([]byte(svg)); err != nil {
		say.Info("cannot handle request: cannot write image: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func labelName(c *semix.Concept) string {
	name := c.ShortName()
	if len(name) <= 10 {
		return name
	}
	for i := len(name) / 2; i < len(name); i++ {
		if name[i] == ' ' {
			return name[:i] + "\n" + name[i+1:]
		}
	}
	return name
}

//w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
func internalError(err error) status {
	return status{err, http.StatusInternalServerError}
}

func badRequest(err error) status {
	return status{err, http.StatusBadRequest}
}

func ok() status {
	return status{nil, http.StatusOK}
}
