package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"bitbucket.org/fflo/semix/pkg/rest"
	"bitbucket.org/fflo/semix/pkg/semix"
)

// M is the map of data for the templates.
type M map[string]interface{}

// Config is the configuration data.
type Config struct {
	Self, Semixd string
}

var (
	infotmpl   *template.Template
	puttmpl    *template.Template
	indextmpl  *template.Template
	gettmpl    *template.Template
	ctxtmpl    *template.Template
	searchtmpl *template.Template
	dir        string
	host       string
	daemon     string
	help       bool
)

func init() {
	flag.StringVar(&dir, "dir", "cmd/semix-httpd/html", "set template directory")
	flag.StringVar(&host, "host", "localhost:8181", "set listen host")
	flag.StringVar(&daemon, "daemon", "localhost:6660", "set host of rest service")
	flag.BoolVar(&help, "help", false, "print this help")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	infotmpl = template.Must(template.ParseFiles(filepath.Join(dir, "info.html")))
	puttmpl = template.Must(template.ParseFiles(filepath.Join(dir, "put.html")))
	indextmpl = template.Must(template.ParseFiles(filepath.Join(dir, "index.html")))
	gettmpl = template.Must(template.ParseFiles(filepath.Join(dir, "get.html")))
	ctxtmpl = template.Must(template.ParseFiles(filepath.Join(dir, "ctx.html")))
	searchtmpl = template.Must(template.ParseFiles(filepath.Join(dir, "search.html")))
	http.HandleFunc("/", rest.WithLogging(rest.WithGet(handle(home))))
	http.HandleFunc("/index", rest.WithLogging(rest.WithGet(handle(home))))
	http.HandleFunc("/info", rest.WithLogging(rest.WithGet(handle(info))))
	http.HandleFunc("/get", rest.WithLogging(rest.WithGet(handle(get))))
	http.HandleFunc("/search", rest.WithLogging(rest.WithGet(handle(search))))
	http.HandleFunc("/ctx", rest.WithLogging(rest.WithGet(handle(ctx))))
	http.HandleFunc("/put", rest.WithLogging(rest.WithGetOrPost(handle(put))))
	http.HandleFunc("/parents", rest.WithLogging(rest.WithGet(handle(parents))))
	http.HandleFunc("/favicon.ico", rest.WithLogging(rest.WithGet(favicon)))
	http.HandleFunc("/js/semix.js", rest.WithLogging(rest.WithGet(semixJS)))
	log.Printf("starting the server on %s", host)
	log.Fatal(http.ListenAndServe(host, nil))
}

type status struct {
	err    error
	status int
}

func handle(f func(*http.Request) (*template.Template, interface{}, status)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t, x, s := f(r)
		if s.err != nil {
			log.Printf("could not handle request: %v", s.err)
			http.Error(w, s.err.Error(), s.status)
			return
		}
		buffer := new(bytes.Buffer)
		if err := t.Execute(buffer, x); err != nil {
			log.Printf("could not execute template: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(s.status)
		w.Header()["Content-Type"] = []string{"text/html; charset=utf-8"}
		if _, err := w.Write(buffer.Bytes()); err != nil {
			log.Printf("could not write html: %v", err)
		}
	}
}

func favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(dir, "favicon.ico"))
}

func semixJS(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(dir, "js", "semix.js"))
}

func search(r *http.Request) (*template.Template, interface{}, status) {
	q := r.URL.Query().Get("q")
	cs, err := rest.NewClient(daemon).Search(q)
	if err != nil {
		return nil, nil, internalError(err)
	}
	return searchtmpl, struct {
		Title    string
		Concepts []*semix.Concept
	}{fmt.Sprintf("%q", q), cs}, ok()
}

func parents(r *http.Request) (*template.Template, interface{}, status) {
	q := r.URL.Query().Get("url")
	cs, err := rest.NewClient(daemon).ParentsURL(q)
	if err != nil {
		return nil, nil, internalError(err)
	}
	return searchtmpl, struct {
		Title    string
		Concepts []*semix.Concept
	}{fmt.Sprintf("parents of %q", q), cs}, ok()
}

func info(r *http.Request) (*template.Template, interface{}, status) {
	q := r.URL.Query().Get("url")
	info, err := rest.NewClient(daemon).InfoURL(q)
	if err != nil {
		return nil, nil, internalError(err)
	}
	return infotmpl, info, ok()
}

func home(r *http.Request) (*template.Template, interface{}, status) {
	return indextmpl, nil, ok()
}

func get(r *http.Request) (*template.Template, interface{}, status) {
	q := r.URL.Query().Get("q")
	ts, err := rest.NewClient(daemon).Get(q)
	if err != nil {
		return nil, nil, internalError(err)
	}
	return gettmpl, struct {
		Query  string
		Tokens rest.Tokens
	}{q, ts}, ok()
}

func ctx(r *http.Request) (*template.Template, interface{}, status) {
	var ctx rest.Context
	var data struct {
		URL     string
		B, E, N int
	}
	if err := rest.DecodeQuery(r.URL.Query(), &data); err != nil {
		return nil, nil, internalError(err)
	}
	ctx, err := rest.NewClient(daemon).Ctx(data.URL, data.B, data.E, data.N)
	if err != nil {
		return nil, nil, internalError(err)
	}
	return ctxtmpl, ctx, ok()
}

func put(r *http.Request) (*template.Template, interface{}, status) {
	switch r.Method {
	case http.MethodPost:
		ct := "text/plain"
		ts, err := rest.NewClient(daemon).PutContent(r.Body, ct, nil, nil)
		if err != nil {
			return nil, nil, internalError(err)
		}
		return puttmpl, ts, ok()
	case http.MethodGet:
		var ps struct {
			URL string
			Ls  []int
			Rs  []string
		}
		if err := rest.DecodeQuery(r.URL.Query(), &ps); err != nil {
			return nil, nil, internalError(err)
		}
		ts, err := rest.NewClient(daemon).PutURL(ps.URL, ps.Ls, nil)
		if err != nil {
			return nil, nil, internalError(err)
		}
		return puttmpl, ts, ok()
	}
	return nil, nil, status{
		fmt.Errorf("invalid request method: %s", r.Method),
		http.StatusBadRequest,
	}
}

func internalError(err error) status {
	return status{err, http.StatusInternalServerError}
}

func ok() status {
	return status{nil, http.StatusOK}
}
