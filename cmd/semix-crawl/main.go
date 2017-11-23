package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/resources"
	"bitbucket.org/fflo/semix/pkg/semix"
	"golang.org/x/net/html"
)

var (
	dir       string
	conf      string
	allowed   *regexp.Regexp
	forbidden *regexp.Regexp
	help      bool
	max       int
	maxjobs   int
	links     chan *url.URL
	crawls    chan semix.Document
	cond      *sync.Cond
	njobs     int
)

func init() {
	var a, f string
	flag.StringVar(&a, "allow", "^https?://de.wikipedia.org/wiki", "allowed url regexp")
	flag.StringVar(&f, "deny", "Datei:", "forbidden url regexp")
	flag.StringVar(&dir, "dir", filepath.Join(os.Getenv("HOME"), "semix"), "semix index directory")
	flag.StringVar(&conf, "resources", "semix.toml", "resources file")
	flag.BoolVar(&help, "help", false, "print help")
	flag.IntVar(&max, "max", 10, "max number of documents to process")
	flag.IntVar(&maxjobs, "jobs", 100, "number of jobs")
	forbidden = regexp.MustCompile(f)
	allowed = regexp.MustCompile(a)
	njobs = 0
	links = make(chan *url.URL, 1000)
	crawls = make(chan semix.Document)
	cond = sync.NewCond(&sync.Mutex{})
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	run(os.Args[len(os.Args)-flag.NArg():])
}

func run(args []string) {
	r, err := resources.Parse(conf)
	if err != nil {
		log.Fatal(err)
	}
	idx, err := index.New(dir, index.DefaultBufferSize)
	if err != nil {
		log.Fatal(err)
	}
	defer idx.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream := index.Put(ctx, idx, semix.Filter(ctx,
		semix.Match(ctx, semix.DFAMatcher{DFA: semix.NewDFA(r.Dictionary, r.Graph)},
			semix.Normalize(ctx, crawl(ctx)))))
	for _, arg := range args {
		url, err := url.Parse(arg)
		if err != nil {
			log.Printf("invalid url %s: %s", url, err)
			continue
		}
		links <- url
	}
	var tokens int
	for t := range stream {
		if t.Err != nil {
			log.Printf("error: %s", t.Err)
		} else {
			tokens++
		}
	}
	log.Printf("indexed %d tokens", tokens)
}

type webpage struct {
	url     *url.URL
	content string
}

func crawl(ctx context.Context) semix.Stream {
	cstream := make(chan semix.StreamToken)
	go func() {
		cctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		defer close(cstream)
		go dispatcher(cctx)
		var n int
		for n < max {
			select {
			case <-ctx.Done():
				return
			case doc := <-crawls:
				log.Printf("[%d] crawled page: %s", n+1, doc.Path())
				select {
				case <-ctx.Done():
					return
				case cstream <- semix.ReadStreamToken(doc):
					n++
				}
			}
		}
	}()
	return cstream
}

func dispatcher(ctx context.Context) {
	urlset := make(map[string]bool)
loop:
	for {
		select {
		case <-ctx.Done():
			return
		case url := <-links:
			if !shouldHandleURL(urlset, url) {
				continue loop
			}
			go handle(ctx, url)
		}
	}
}

func shouldHandleURL(urlset map[string]bool, url *url.URL) bool {
	str := url.String()
	if urlset[str] {
		return false
	}
	urlset[str] = true
	if allowed.FindString(str) == "" {
		return false
	}
	if forbidden.FindString(str) != "" {
		return false
	}
	return true
}

func handle(ctx context.Context, url *url.URL) {
	id := startJob()
	defer stopJob()
	ms := time.Duration(rand.Intn(500))
	time.Sleep(time.Duration(ms * time.Millisecond))
	log.Printf("[%d] sending request: [GET] %s", id, url)
	res, err := http.Get(url.String())
	if err != nil {
		log.Printf("error: [GET] %s: %s", url, err)
		return
	}
	defer res.Body.Close()
	if !strings.Contains(res.Header.Get("Content-Type"), "text/html") {
		// log.Printf("skipping Content-Type: %q", res.Header.Get("Content-Type"))
		return
	}
	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, res.Body); err != nil {
		log.Printf("could not copy content of %s: %s", url, err)
		return
	}
	content := buffer.String()
	go dispatchLinks(ctx, content, url)
	select {
	case <-ctx.Done():
	case crawls <- semix.NewHTMLDocument(url.String(), strings.NewReader(content)):
	}
}

func startJob() int {
	cond.L.Lock()
	for njobs >= maxjobs {
		cond.Wait()
	}
	defer cond.L.Unlock()
	if njobs >= maxjobs {
		panic("what?")
	}
	njobs++
	return njobs
}

func stopJob() {
	cond.L.Lock()
	defer cond.L.Unlock()
	njobs--
}

func dispatchLinks(ctx context.Context, content string, base *url.URL) {
	go func() {
		doc, err := html.Parse(strings.NewReader(content))
		if err != nil {
			log.Printf("error parsing html: %s", err)
		}
		var f func(*html.Node)
		f = func(node *html.Node) {
			if node.Type == html.ElementNode && node.Data == "a" {
				for _, a := range node.Attr {
					if a.Key == "href" {
						link, err := makeURL(base, a.Val)
						if err != nil {
							log.Printf("invalid url %s %s: %s", base, a.Val, err)
							break
						}
						select {
						case <-ctx.Done():
							return
						case links <- link:
						}
						break
					}
				}
			}
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(doc)
	}()
}

func makeURL(base *url.URL, link string) (*url.URL, error) {
	l, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if l.IsAbs() {
		return l, nil
	}
	// log.Printf("%s://%s/", base.Scheme, base.Host)
	// panic("")
	return url.Parse(fmt.Sprintf("%s://%s%s", base.Scheme, base.Host, link))
}
