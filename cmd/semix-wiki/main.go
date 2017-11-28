package main

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
)

type matches []*regexp.Regexp

func (m matches) String() string {
	var strs []string
	for _, re := range m {
		strs = append(strs, re.String())
	}
	return strings.Join(strs, ",")
}

func (m *matches) Set(val string) error {
	re, err := regexp.Compile(val)
	if err != nil {
		return err
	}
	*m = append(*m, re)
	return nil
}

func (m matches) Match(article string) bool {
	if len(m) == 0 {
		return true
	}
	for _, re := range m {
		if re.FindString(article) != "" {
			return true
		}
	}
	return false
}

var (
	ms matches
)

func init() {
	flag.Var(&ms, "m", "list of regexes to match articles")
}

func main() {
	flag.Parse()
	fmt.Printf("matches: %s\n", ms)
}
