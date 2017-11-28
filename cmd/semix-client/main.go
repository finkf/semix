package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"bitbucket.org/fflo/semix/pkg/rest"
	"bitbucket.org/fflo/semix/pkg/semix"
)

var (
	daemon     string
	search     string
	predicates string
	put        string
	get        string
	url        string
	id         int
	info       bool
	parents    bool
	help       bool
	client     rest.Client
	ls         levs
	rs         resolvers
)

func init() {
	flag.StringVar(&daemon, "daemon", "http://localhost:6660", "set daemon host")
	flag.StringVar(&search, "search", "", "search for concepts")
	flag.StringVar(&predicates, "predicates", "", "search for predicates")
	flag.StringVar(&put, "put", "", "put files or directories into the index")
	flag.StringVar(&get, "get", "", "execute a query on the index")
	flag.IntVar(&id, "id", 0, "set search ID")
	flag.StringVar(&url, "url", "", "set search URL")
	flag.BoolVar(&info, "info", false, "get info (needs -id or -url)")
	flag.BoolVar(&parents, "parents", false, "get parents of concept (needs -id or -url)")
	flag.BoolVar(&help, "help", false, "print this help")
	flag.Var(&rs, "r", "add named resolver (can be set multiple times)")
	flag.Var(&ls, "l", "add levenshtein distance for approximate search (can be set multiple times)")
}

func main() {
	flag.Parse()
	if help {
		flag.PrintDefaults()
		return
	}
	client = rest.NewClient(daemon)
	if search != "" {
		doSearch()
	}
	if predicates != "" {
		doPredicates()
	}
	if info {
		doInfo()
	}
	if parents {
		doParents()
	}
	if get != "" {
		doGet()
	}
	if put != "" {
		doPut()
	}
}

func doParents() {
	assertSearchOK()
	var err error
	var cs []*semix.Concept
	if url != "" {
		cs, err = client.ParentsURL(url)
	}
	if id != 0 {
		cs, err = client.ParentsID(id)
	}
	if err != nil {
		log.Fatal(err)
	}
	printConcepts(cs)
}

func doInfo() {
	assertSearchOK()
	var err error
	var info rest.ConceptInfo
	if url != "" {
		info, err = client.InfoURL(url)
	}
	if id != 0 {
		info, err = client.InfoID(id)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", info)
}

func doSearch() {
	cs, err := client.Search(search)
	if err != nil {
		log.Fatal(err)
	}
	printConcepts(cs)
}

func doPredicates() {
	cs, err := client.Predicates(predicates)
	if err != nil {
		log.Fatal(err)
	}
	printConcepts(cs)
}

func doGet() {
	ts, err := client.Get(get)
	if err != nil {
		log.Fatal(err)
	}
	printTokens(ts)
}

func doPut() {
	ms, err := filepath.Glob(put)
	if err != nil {
		log.Fatal(err)
	}
	for _, path := range ms {
		putFileOrDir(path)
	}
}

func putFileOrDir(path string) {
	info, err := os.Lstat(path)
	if err != nil {
		log.Fatal(err)
	}
	if info.IsDir() {
		putDir(path)
	}
	if info.Mode().IsRegular() {
		putFile(path)
	}
}

func putDir(path string) {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if info.IsDir() {
			putDir(path)
		}
		if info.Mode().IsRegular() {
			putFile(path)
		}
		return nil
	})
}

func putFile(path string) {
	var ts rest.Tokens
	var err error
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		ts, err = client.PutURL(path, ls, rs)
	} else {
		is, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer is.Close()
		ts, err = client.PutContent(is, "text/plain", ls, rs)
	}
	if err != nil {
		log.Fatal(err)
	}
	printTokens(ts)
}

func printTokens(ts rest.Tokens) {
	for _, t := range ts.Tokens {
		fmt.Printf("%v\n", t)
	}
}

func printConcepts(cs []*semix.Concept) {
	for _, c := range cs {
		fmt.Printf("%s\n", c)
	}
}

func assertSearchOK() {
	if id == 0 && url == "" {
		log.Fatal("missing concept id or url")
	}
}

type levs []int

func (ls levs) String() string {
	var strs []string
	for _, l := range ls {
		strs = append(strs, fmt.Sprintf("%d", l))
	}
	return strings.Join(strs, ",")
}

func (ls *levs) Set(val string) error {
	l, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	*ls = append(*ls, l)
	return nil
}

type resolvers []string

func (rs resolvers) String() string {
	return strings.Join(rs, ",")
}

func (rs *resolvers) Set(val string) error {
	*rs = append(*rs, val)
	return nil
}
