package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"bitbucket.org/fflo/semix/pkg/restd"
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
	restHost   string
	help       bool
)

func init() {
	flag.StringVar(&dir, "dir", "cmd/semix-httpd/html", "set template directory")
	flag.StringVar(&host, "host", "localhost:8181", "set listen host")
	flag.StringVar(&restHost, "restd", "localhost:6060", "set host of rest service")
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
	http.HandleFunc("/", withLogging(withGet(handle(home))))
	http.HandleFunc("/index", withLogging(withGet(handle(home))))
	http.HandleFunc("/info", withLogging(withGet(handle(info))))
	http.HandleFunc("/get", withLogging(withGet(handle(get))))
	http.HandleFunc("/search", withLogging(withGet(handle(search))))
	http.HandleFunc("/ctx", withLogging(withGet(handle(ctx))))
	http.HandleFunc("/put", withLogging(handle(put)))
	log.Printf("starting the server")
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

func withLogging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("handling request for [%s] %s", r.Method, r.RequestURI)
		f(w, r)
		log.Printf("handled request for [%s] %s", r.Method, r.RequestURI)
	}
}

func withGet(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("invalid request method: %v", r.Method)
			http.Error(w, fmt.Sprintf("invalid request method: %s", r.Method), http.StatusBadRequest)
			return
		}
		f(w, r)
	}
}

func search(r *http.Request) (*template.Template, interface{}, status) {
	var cs []semix.Concept
	q := r.URL.Query().Get("q")
	if err := semixdGet(fmt.Sprintf("/search?q=%s", url.QueryEscape(q)), &cs); err != nil {
		return nil, nil, internalError(err)
	}
	return searchtmpl, struct{ Concepts []semix.Concept }{cs}, ok()
}

func info(r *http.Request) (*template.Template, interface{}, status) {
	q := r.URL.Query().Get("q")
	var info restd.ConceptInfo
	if err := semixdGet(fmt.Sprintf("/info?q=%s", url.QueryEscape(q)), &info); err != nil {
		return nil, nil, internalError(err)
	}
	return infotmpl, info, ok()
}

func home(r *http.Request) (*template.Template, interface{}, status) {
	return indextmpl, nil, ok()
}

func get(r *http.Request) (*template.Template, interface{}, status) {
	q := r.URL.Query().Get("q")
	var ts restd.Tokens
	if err := semixdGet(fmt.Sprintf("/get?q=%s", url.QueryEscape(q)), &ts); err != nil {
		return nil, nil, internalError(err)
	}
	return gettmpl, ts, ok()
}

func ctx(r *http.Request) (*template.Template, interface{}, status) {
	var ctx restd.Context
	url := fmt.Sprintf("/ctx?url=%s&b=%s&e=%s&n=%s",
		url.QueryEscape(r.URL.Query().Get("url")),
		url.QueryEscape(r.URL.Query().Get("b")),
		url.QueryEscape(r.URL.Query().Get("e")),
		url.QueryEscape(r.URL.Query().Get("n")),
	)
	if err := semixdGet(url, &ctx); err != nil {
		return nil, nil, internalError(err)
	}
	return ctxtmpl, ctx, ok()
}

func put(r *http.Request) (*template.Template, interface{}, status) {
	switch r.Method {
	case "POST":
		return putPost(r)
	case "GET":
		return putGet(r)
	default:
		return nil, nil, status{
			fmt.Errorf("invalid request method: %s", r.Method),
			http.StatusBadRequest,
		}
	}
}

func putGet(r *http.Request) (*template.Template, interface{}, status) {
	q := r.URL.Query().Get("url")
	var info restd.Tokens
	if err := semixdGet(fmt.Sprintf("/put?url=%s", url.QueryEscape(q)), &info); err != nil {
		return nil, nil, internalError(err)
	}
	return puttmpl, info, ok()
}

func putPost(r *http.Request) (*template.Template, interface{}, status) {
	var info restd.Tokens
	ctype := "text/plain"
	if len(r.Header["Content-Type"]) > 0 {
		ctype = strings.Join(r.Header["Content-Type"], ",")
	}
	if err := semixdPost("/put", ctype, r.Body, &info); err != nil {
		return nil, nil, internalError(err)
	}
	return puttmpl, info, ok()
}

func semixdPost(path string, ctype string, r io.Reader, data interface{}) error {
	url := "http://localhost:6060" + path
	log.Printf("sending: [POST] %s", url)
	res, err := http.Post(url, ctype, r)
	if err != nil {
		return fmt.Errorf("could not [POST] %s: %v", url, err)
	}
	defer res.Body.Close()
	log.Printf("response: [POST] %s: %s", url, res.Status)
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return fmt.Errorf("invalid response code [POST] %s: %s", url, res.Status)
	}
	err = json.NewDecoder(res.Body).Decode(data)
	if err != nil {
		return fmt.Errorf("could not decode response: %v", err)
	}
	return nil
}

func semixdGet(path string, data interface{}) error {
	url := "http://localhost:6060" + path
	log.Printf("sending: [GET] %s", url)
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("could not [GET] %s: %v", url, err)
	}
	defer res.Body.Close()
	log.Printf("response: [GET] %s: %s", url, res.Status)
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return fmt.Errorf("invalid response code [GET] %s: %s", url, res.Status)
	}
	err = json.NewDecoder(res.Body).Decode(data)
	if err != nil {
		return fmt.Errorf("could not decode response: %v", err)
	}
	return nil
}

func internalError(err error) status {
	return status{err, http.StatusInternalServerError}
}

func ok() status {
	return status{nil, http.StatusOK}
}
