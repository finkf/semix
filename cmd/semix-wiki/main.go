package main

import (
	"compress/bzip2"
	"encoding/xml"
	"flag"
	"html"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"bitbucket.org/fflo/semix/pkg/args"
	"bitbucket.org/fflo/semix/pkg/rest"
)

type article struct {
	name, content string
}

var (
	articles  args.RegexList
	file      string
	daemon    string
	worker    int
	narticles int
)

func init() {
	flag.Var(&articles, "a", "list of regexes to match articles")
	flag.IntVar(&worker, "w", 2, "number of workers")
	flag.IntVar(&narticles, "n", 0, "maximal number of articles")
	flag.StringVar(&daemon, "daemon", "http://localhost:6606", "set address of daemon")
	flag.StringVar(&file, "f", "", "set input file")
}

func main() {
	flag.Parse()
	achan := make(chan article, worker)
	var wg sync.WaitGroup
	wg.Add(worker + 1)
	go readWiki(file, &wg, achan)
	for i := 0; i < worker; i++ {
		go work(i, &wg, achan)
	}
	wg.Wait()
}

func work(i int, wg *sync.WaitGroup, achan <-chan article) {
	client := rest.NewClient(daemon)
	defer wg.Done()
	for article := range achan {
		log.Printf("%s\n%s", article.name, article.content)
		continue
		if es, err := client.PutContent(strings.NewReader(article.content),
			article.name, "text/plain", nil, nil); err != nil {
			log.Printf("[error] %s", err)
		} else {
			log.Printf("[%d] put article: %s (%d tokens)", i, article.name, len(es))
		}
	}
}

func readWiki(path string, wg *sync.WaitGroup, achan chan article) {
	defer close(achan)
	defer wg.Done()
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	d := xml.NewDecoder(bzip2.NewReader(file))
	var inpage, intitle, intext bool
	var article article
	var n int
	for narticles == 0 || n < narticles {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "page":
				inpage = true
			case "text":
				intext = inpage
			case "title":
				intitle = inpage
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "page":
				inpage = false
			case "text":
				intext = false
			case "title":
				intitle = false
			}
		case xml.CharData:
			if intitle {
				article.name = string(t)
			}
			if intext {
				article.content = cleanup(string(t))
				achan <- article
				n++
			}
		}
	}
}

func cleanup(content string) string {
	content = html.UnescapeString(content)
	content = regexp.MustCompile("(?si)<ref.*?>.*?</ref>").ReplaceAllLiteralString(content, " ")
	content = regexp.MustCompile("(?si)<ref.*?/>").ReplaceAllLiteralString(content, " ")
	content = regexp.MustCompile("(?si)<!--.*?-->").ReplaceAllLiteralString(content, " ")
	content = regexp.MustCompile("(?si)<br.*?/?>").ReplaceAllLiteralString(content, " ")
	content = regexp.MustCompile("(?si)<u>(.*?)</u>").ReplaceAllString(content, "$1")
	content = regexp.MustCompile("(?si)<sub>(.*?)</sub>").ReplaceAllString(content, "$1")
	content = regexp.MustCompile("(?si)<small>(.*?)</small>").ReplaceAllString(content, "$1")
	content = regexp.MustCompile("(?si)<math>.*?</math>").ReplaceAllLiteralString(content, " ")
	content = regexp.MustCompile("(?si){{.*?}}").ReplaceAllLiteralString(content, " ")
	content = regexp.MustCompile("(?si){\\|.*?\\|}").ReplaceAllLiteralString(content, " ")
	content = regexp.MustCompile("(?si)https?://\\S*").ReplaceAllLiteralString(content, " ")
	content = regexp.MustCompile("(?si)==\\s*Literatur\\s*==.*").ReplaceAllLiteralString(content, " ")
	content = regexp.MustCompile("(?si)==\\s*Weblinks\\s*==.*").ReplaceAllLiteralString(content, " ")
	return content
}

func match(article string) bool {
	if len(articles) == 0 {
		return true
	}
	for _, re := range articles {
		if re.FindString(article) != "" {
			return true
		}
	}
	return false
}
