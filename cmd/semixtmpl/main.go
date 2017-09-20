package main

import (
	"encoding/json"
	"fmt"
	"html/template"
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

var tmpl *template.Template
var client *http.Client

func main() {
	tmpl = template.Must(template.ParseFiles("cmd/semixtmpl/tmpls/info.html"))
	client = &http.Client{}
	http.HandleFunc("/info", handle)
	log.Fatalf(http.ListenAndServe(":8080", nil).Error())
}

func handle(w http.ResponseWriter, req *http.Request) {
	log.Printf("serving request for %s", req.RequestURI)
	if req.Method != "GET" {
		log.Printf("invalid method: %s", req.Method)
		http.Error(w, "not a GET request", http.StatusBadRequest)
		return
	}
	q := req.URL.Query()["q"]
	if len(q) != 1 {
		log.Printf("invalid query: %v", q)
		http.Error(w, "invalid query paramters", http.StatusBadRequest)
		return
	}
	info, err := get(q[0])
	if err != nil {
		log.Printf("could not load info: %v", err)
		http.Error(w, "could not find info", http.StatusNotFound)
	}
	if err := tmpl.Execute(w, info); err != nil {
		log.Printf("could not load info: %v", err)
		http.Error(w, "could not load info: %v", http.StatusInternalServerError)
	}
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
