package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"bitbucket.org/fflo/semix/pkg/net"
)

// LookupInfo is some info
type LookupInfo struct {
	Query   string
	Subject string
	Links   map[string][]string
	Entries []string
}

// M is the map of data for the templates.
type M map[string]interface{}

// Config is the configuration data.
type Config struct {
	Self, Semixd string
}

var (
	infotmpl  *template.Template
	puttmpl   *template.Template
	indextmpl *template.Template
	gettmpl   *template.Template
	config    Config
)

func main() {
	config.Self = "http://localhost:8080"
	config.Semixd = "http://localhost:6060"
	infotmpl = template.Must(template.ParseFiles("cmd/semixtmpl/tmpls/info.html"))
	puttmpl = template.Must(template.ParseFiles("cmd/semixtmpl/tmpls/put.html"))
	indextmpl = template.Must(template.ParseFiles("cmd/semixtmpl/tmpls/index.html"))
	gettmpl = template.Must(template.ParseFiles("cmd/semixtmpl/tmpls/get.html"))
	http.HandleFunc("/", requestFunc(index))
	http.HandleFunc("/index", requestFunc(index))
	http.HandleFunc("/info", requestFunc(info))
	http.HandleFunc("/put", requestFunc(put))
	http.HandleFunc("/get", requestFunc(get))
	log.Printf("starting the server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func requestFunc(h func(*http.Request) ([]byte, int, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, status, err := h(r)
		if err != nil {
			log.Printf("error: %v", err)
			w.Header()["Content-Type"] = []string{"text/plain", "charset=utf-8"}
			http.Error(w, err.Error(), status)
			return
		}
		w.WriteHeader(status)
		w.Header()["Content-Type"] = []string{"text/html", "charset=utf-8"}
		if _, err := w.Write(data); err != nil {
			log.Printf("could not write response: %v", err)
		}
	}
}

func info(r *http.Request) ([]byte, int, error) {
	log.Printf("serving request for %s", r.RequestURI)
	if r.Method != "GET" {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid request method: %s", r.Method)
	}
	q := r.URL.Query()["q"]
	if len(q) != 1 {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query q=%v", q)
	}
	var info net.ConceptInfo
	err := semixdGet(fmt.Sprintf("/search?q=%s", url.QueryEscape(q[0])), &info)
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not talk to semixd: %v", err)
	}
	buffer := new(bytes.Buffer)
	if err := infotmpl.Execute(buffer, info); err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not write html: %v", err)
	}
	log.Printf("served request for %s", r.RequestURI)
	return buffer.Bytes(), http.StatusOK, nil
}

func index(r *http.Request) ([]byte, int, error) {
	log.Printf("serving request for %s", r.RequestURI)
	if r.Method != "GET" {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid request method: %s", r.Method)
	}
	buffer := new(bytes.Buffer)
	if err := indextmpl.Execute(buffer, M{"config": config}); err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not write html: %v", err)
	}
	log.Printf("served request for %s", r.RequestURI)
	return buffer.Bytes(), http.StatusOK, nil
}

func get(r *http.Request) ([]byte, int, error) {
	log.Printf("serving request for %s", r.RequestURI)
	if r.Method != http.MethodGet {
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid request method %s", r.Method)
	}
	q := r.URL.Query()["q"]
	if len(q) != 1 {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query parameter q=%v", q)
	}
	var ts net.Tokens
	err := semixdGet(fmt.Sprintf("/get?q=%s", url.QueryEscape(q[0])), &ts)
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not talk to semixd: %v", err)
	}
	buffer := new(bytes.Buffer)
	if err := gettmpl.Execute(buffer, ts); err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not write html: %v", err)
	}
	log.Printf("served request for %s", r.RequestURI)
	return buffer.Bytes(), http.StatusOK, nil
}

func put(r *http.Request) ([]byte, int, error) {
	log.Printf("serving request for %s", r.RequestURI)
	switch r.Method {
	case "POST":
		return putPost(r)
	case "GET":
		return putGet(r)
	default:
		return nil, http.StatusForbidden,
			fmt.Errorf("invalid request method: %s", r.Method)
	}
}

func putGet(r *http.Request) ([]byte, int, error) {
	q := r.URL.Query()["url"]
	if len(q) != 1 {
		return nil, http.StatusBadRequest,
			fmt.Errorf("invalid query parameter url=%v", q)
	}
	var info net.Tokens
	err := semixdGet(fmt.Sprintf("/put?url=%s", url.QueryEscape(q[0])), &info)
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not talk to semixd: %v", err)
	}
	data := struct {
		Config Config
		Data   net.Tokens
	}{
		Config: config,
		Data:   info,
	}
	buffer := new(bytes.Buffer)
	if err := puttmpl.Execute(buffer, data); err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not write html: %v", err)
	}
	log.Printf("served request for %s", r.RequestURI)
	return buffer.Bytes(), http.StatusOK, nil
}

func putPost(r *http.Request) ([]byte, int, error) {
	var info net.Tokens
	ctype := "text/plain"
	if len(r.Header["Content-Type"]) > 0 {
		ctype = strings.Join(r.Header["Content-Type"], ",")
	}
	err := semixdPost("/put", ctype, r.Body, &info)
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not talk to semixd: %v", err)
	}
	data := struct {
		Config Config
		Data   net.Tokens
	}{
		Config: config,
		Data:   info,
	}
	buffer := new(bytes.Buffer)
	if err := puttmpl.Execute(buffer, data); err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not write html: %v", err)
	}
	log.Printf("served request for %s", r.RequestURI)
	return buffer.Bytes(), http.StatusOK, nil
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
