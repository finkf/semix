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
	http.HandleFunc("/", requestFunc(index))
	http.HandleFunc("/info", requestFunc(info))
	http.HandleFunc("/put", requestFunc(put))
	log.Fatalf(http.ListenAndServe(":8080", nil).Error())
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
	info, err := get(q[0])
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
	path := config.Semixd + fmt.Sprintf("/put?url=%s", url.QueryEscape(q[0]))
	res, err := http.Get(path)
	if err != nil || res.StatusCode != 200 {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not talk to semixd: %v", err)
	}
	defer res.Body.Close()
	var info IndexInfo
	d := json.NewDecoder(res.Body)
	if err := d.Decode(&info); err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not decode json: %v", err)
	}
	counts := getCounts(info)
	buffer := new(bytes.Buffer)
	if err := puttmpl.Execute(buffer, M{"config": config, "data": info, "counts": counts}); err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not write html: %v", err)
	}
	log.Printf("served request for %s", r.RequestURI)
	return buffer.Bytes(), http.StatusOK, nil
}

func putPost(r *http.Request) ([]byte, int, error) {
	info, err := post(r.Body)
	if err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not talk to semixd: %v", err)
	}
	buffer := new(bytes.Buffer)
	if err := puttmpl.Execute(buffer, info); err != nil {
		return nil, http.StatusInternalServerError,
			fmt.Errorf("could not write html: %v", err)
	}
	log.Printf("served request for %s", r.RequestURI)
	return buffer.Bytes(), http.StatusOK, nil
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
