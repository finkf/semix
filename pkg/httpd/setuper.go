package httpd

import (
	"html/template"
	"path/filepath"

	"github.com/finkf/semix/pkg/say"
)

var (
	funcs = template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}
)

type setuper struct {
	err error
}

func (s *setuper) setup(server *Server) error {
	server.infotmpl = s.setupTemplate(filepath.Join(server.dir, "info.html"))
	server.puttmpl = s.setupTemplate(filepath.Join(server.dir, "put.html"))
	server.indextmpl = s.setupTemplate(filepath.Join(server.dir, "index.html"))
	server.ctxtmpl = s.setupTemplate(filepath.Join(server.dir, "ctx.html"))
	server.searchtmpl = s.setupTemplate(filepath.Join(server.dir, "search.html"))
	server.gettmpl = s.setupTemplate(filepath.Join(server.dir, "get.html"))
	server.dumptmpl = s.setupTemplate(filepath.Join(server.dir, "dump.html"))
	server.setuptmpl = s.setupTemplate(filepath.Join(server.dir, "setup.html"))
	return s.err
}

func (s *setuper) setupTemplate(tmpl string) *template.Template {
	if s.err != nil {
		return nil
	}
	say.Debug("setting up template %s", tmpl)
	t, err := template.New(filepath.Base(tmpl)).Funcs(funcs).ParseFiles(tmpl)
	if err != nil {
		s.err = err
		return nil
	}
	return t
}
