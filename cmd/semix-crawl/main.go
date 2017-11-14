package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"bitbucket.org/fflo/semix/pkg/config"
	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/semix"
	"golang.org/x/net/html"
)

var (
	dir        string
	configFile string
	prefix     *regexp.Regexp
	help       bool
	max        int
)

func init() {
	var pre string
	flag.StringVar(&pre, "prefix", "^https?://de.wikipedia.org/wiki", "allowed url prefix regexp")
	prefix = regexp.MustCompile(pre)
	flag.StringVar(&dir, "dir", filepath.Join(os.Getenv("HOME"), "semix"), "semix index directory")
	flag.StringVar(&configFile, "config", "semix.toml", "configuration file")
	flag.BoolVar(&help, "help", false, "print help")
	flag.IntVar(&max, "max", 1000, "max number of documents to process")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	run(os.Args[len(os.Args)-flag.NArg():])
}

var (
	dispatchc chan *url.URL
	indexerc  chan webpage
)

func run(args []string) {
	g, d, err := config.Parse(configFile)
	if err != nil {
		log.Fatal(err)
	}
	idx, err := index.New(dir, index.DefaultBufferSize)
	if err != nil {
		log.Fatal(err)
	}
	dispatchc = make(chan *url.URL, len(args))
	indexerc = make(chan webpage)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream := index.Put(ctx, idx, semix.Filter(ctx,
		semix.Match(ctx, semix.DFAMatcher{DFA: semix.NewDFA(d, g)},
			semix.Normalize(ctx, crawl(ctx)))))
	fmt.Printf("Args: %v\n", args)
	for _, arg := range args {
		url, err := url.Parse(arg)
		if err != nil {
			log.Printf("invalid url %s: %s", url, err)
			continue
		}
		dispatchc <- url
	}
	for t := range stream {
		log.Printf("%s", t.Token)
	}
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
			case page := <-indexerc:
				log.Printf("[%d] crawled page: %s", n+1, page.url)
				token := semix.StreamToken{
					Token: semix.Token{
						Token: page.content,
						Begin: 0,
						End:   len(page.content),
						Path:  page.url.String(),
					},
				}
				select {
				case <-ctx.Done():
					return
				case cstream <- token:
					n++
				}
			}
		}
	}()
	return cstream
}

func dispatcher(ctx context.Context) {
	urlset := make(map[string]bool)
	for {
		select {
		case <-ctx.Done():
			return
		case url := <-dispatchc:
			if urlset[url.String()] {
				// log.Printf("skipping %s", url)
				break
			}
			urlset[url.String()] = true
			if prefix.FindString(url.String()) == "" {
				// log.Printf("skipping %s", url)
				break
			}
			go handle(ctx, url)
		}
	}
}

func handle(ctx context.Context, url *url.URL) {
	// log.Printf("sending request: [GET] %s", url)
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
	go dispatchLinks(ctx, bytes.NewBuffer(buffer.Bytes()), url)
	select {
	case <-ctx.Done():
	case indexerc <- webpage{url, buffer.String()}:
	}
}

func dispatchLinks(ctx context.Context, r io.Reader, base *url.URL) {
	go func() {
		doc, err := html.Parse(r)
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
						case dispatchc <- link:
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
