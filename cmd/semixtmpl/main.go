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
	"sort"
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
	client    *http.Client
	config    Config
)

func main() {
	config.Self = "http://localhost:8080"
	config.Semixd = "http://localhost:6060"
	infotmpl = template.Must(template.ParseFiles("cmd/semixtmpl/tmpls/info.html"))
	puttmpl = template.Must(template.ParseFiles("cmd/semixtmpl/tmpls/put.html"))
	indextmpl = template.Must(template.ParseFiles("cmd/semixtmpl/tmpls/index.html"))
	client = &http.Client{}
	http.HandleFunc("/", index)
	http.HandleFunc("/info", requestFunc(info))
	http.HandleFunc("/put", put)
	log.Fatalf(http.ListenAndServe(":8080", nil).Error())
}

func requestFunc(h func(*http.Request) ([]byte, int, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, status, err := h(r)
		if err != nil {
			log.Printf("error: %v", err)
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
	info, err := get(q[0])
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not connect to semixd: %v", err)
	}
	buffer := new(bytes.Buffer)
	if err := infotmpl.Execute(buffer, info); err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not write html: %v", err)
	}
	log.Printf("served request for %s", r.RequestURI)
	return buffer.Bytes(), http.StatusOK, nil
}

func index(w http.ResponseWriter, req *http.Request) {
	log.Printf("serving request for %s", req.RequestURI)
	if req.Method != "GET" {
		log.Printf("invalid method: %s", req.Method)
		http.Error(w, "not a GET request", http.StatusBadRequest)
		return
	}
	if err := indextmpl.Execute(w, M{"config": config}); err != nil {
		log.Printf("could not load info: %v", err)
		http.Error(w, "could not load info: %v", http.StatusInternalServerError)
		return
	}
	log.Printf("served request for %s", req.RequestURI)
}

func put(w http.ResponseWriter, req *http.Request) {
	log.Printf("serving request for %s", req.RequestURI)
	switch req.Method {
	case "POST":
		putPost(w, req)
	case "GET":
		putGet(w, req)
	default:
		log.Printf("invalid method: %s", req.Method)
		http.Error(w, "not a POST or GET request", http.StatusBadRequest)
	}
}

func putGet(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()["url"]
	if len(q) != 1 {
		log.Printf("invalid query parameter url=%v", q)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	path := config.Semixd + fmt.Sprintf("/put?url=%s", url.QueryEscape(q[0]))
	res, err := http.Get(path)
	if err != nil || res.StatusCode != 200 {
		log.Printf("could not get %q: %v (%s)", path, err, res.Status)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()
	var info IndexInfo
	d := json.NewDecoder(res.Body)
	if err := d.Decode(&info); err != nil {
		log.Printf("could not decode json: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	counts := getCounts(info)
	log.Printf("counts: %d", len(counts))
	if err := puttmpl.Execute(w, M{"config": config, "data": info, "counts": counts}); err != nil {
		log.Printf("could not execute template: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	log.Printf("served request for %s", req.RequestURI)
}

func putPost(w http.ResponseWriter, req *http.Request) {
	info, err := post(req.Body)
	if err != nil {
		log.Printf("could not load info: %v", err)
		http.Error(w, "could not load info", http.StatusInternalServerError)
		return
	}
	log.Printf("got %d tokens", len(info.Tokens))
	if err := puttmpl.Execute(w, info); err != nil {
		log.Printf("could not execute template: %v", err)
		http.Error(w, "could nto execute template", http.StatusInternalServerError)
		return
	}
	log.Printf("served request for %s", req.RequestURI)
}

func get(q string) (LookupInfo, error) {
	var info LookupInfo
	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("http://localhost:6060/search?q=%s", url.PathEscape(q)),
		nil)
	if err != nil {
		return info, err
	}
	res, err := client.Do(req)
	if err != nil {
		return info, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return info, fmt.Errorf("invalid response code: %d", res.StatusCode)
	}
	d := json.NewDecoder(res.Body)
	if err := d.Decode(&info); err != nil {
		return info, err
	}
	srt(&info)
	return info, nil
}

func srt(info *LookupInfo) {
	sort.Strings(info.Entries)
	for p := range info.Links {
		sort.Strings(info.Links[p])
	}
}

type CountPair struct {
	Freq float32
	URL  string
}

func getCounts(info IndexInfo) []CountPair {
	counts := make(map[string]int)
	var n int
	for _, t := range info.Tokens {
		n++
		counts[t.ConceptURL]++
		for _, es := range t.Links {
			for _, url := range es {
				counts[url]++
			}
		}
	}
	var list []CountPair
	for url, c := range counts {
		list = append(list, CountPair{URL: url, Freq: float32(c) / float32(n)})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Freq > list[j].Freq
	})
	return list
}

// TokenInfo is used for the templates.
type TokenInfo struct {
	Token, ConceptURL, Path string
	Begin, End              int
	Links                   map[string][]string
}

// IndexInfo is a list of TokenInfo.
type IndexInfo struct {
	Tokens []TokenInfo
}

func post(r io.Reader) (IndexInfo, error) {
	var info IndexInfo
	req, err := http.NewRequest(http.MethodPost, config.Semixd+"/put", r)
	if err != nil {
		return info, err
	}
	res, err := client.Do(req)
	if err != nil {
		return info, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return info, fmt.Errorf("invalid response code: %d", res.StatusCode)
	}
	d := json.NewDecoder(res.Body)
	if err := d.Decode(&info); err != nil {
		return info, err
	}
	return info, nil
}
