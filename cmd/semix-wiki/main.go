package main

import (
	"flag"
	"fmt"

	"bitbucket.org/fflo/semix/pkg/args"
)

var (
	articles args.RegexList
)

func init() {
	flag.Var(&articles, "a", "list of regexes to match articles")
}

func main() {
	flag.Parse()
	fmt.Printf("articles: %s\n", articles)
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
