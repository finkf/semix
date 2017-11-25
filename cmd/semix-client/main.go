package main

import (
	"flag"
	"log"
	"os"
	"text/template"

	"bitbucket.org/fflo/semix/pkg/rest"
)

var (
	format string
	host   string
	search string
)

func init() {
	flag.StringVar(&host, "daemon", "http://localhost:6660", "set daemon host")
	flag.StringVar(&format, "f", "", "set formating template")
	flag.StringVar(&search, "search", "", "set search string")
}

func main() {
	flag.Parse()
	c := rest.NewClient(host)
	cs, err := c.Search(search)
	if err != nil {
		log.Fatal(err)
	}
	t := template.Must(template.New("format").Parse(format))
	if err := t.Execute(os.Stdout, cs); err != nil {
		log.Fatal(err)
	}
}
