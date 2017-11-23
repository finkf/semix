package main

import (
	"flag"
	"log"
	"os"
	"text/template"

	"bitbucket.org/fflo/semix/pkg/restd"
)

var (
	format string
	host   string
	search string
)

func init() {
	flag.StringVar(&host, "host", "http://localhost:6060", "set host")
	flag.StringVar(&format, "f", "", "set formating template")
	flag.StringVar(&search, "search", "", "set search string")
}

func main() {
	flag.Parse()
	c := restd.NewClient(host)
	cs, err := c.Search(search)
	if err != nil {
		log.Fatal(err)
	}
	t := template.Must(template.New("format").Parse(format))
	if err := t.Execute(os.Stdout, cs); err != nil {
		log.Fatal(err)
	}
}
