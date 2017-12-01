package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"bitbucket.org/fflo/semix/pkg/args"
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
	threshold  float64
	memsize    int
	id         int
	filelist   bool
	local      bool
	info       bool
	parents    bool
	help       bool
	client     rest.Client
	ls         args.IntList
	rs         args.StringList
)

func init() {
	flag.Float64Var(&threshold, "threshold", 0, "set threshold for automatic resolver")
	flag.IntVar(&memsize, "memsize", 0, "set memory size for resolvers")
	flag.StringVar(&daemon, "daemon", "http://localhost:6660", "set daemon host")
	flag.StringVar(&search, "search", "", "search for concepts")
	flag.StringVar(&predicates, "predicates", "", "search for predicates")
	flag.StringVar(&put, "put", "", "put files or directories into the index")
	flag.StringVar(&get, "get", "", "execute a query on the index")
	flag.IntVar(&id, "id", 0, "set search ID")
	flag.StringVar(&url, "url", "", "set search URL")
	flag.BoolVar(&filelist, "filelist", false, "treat put arguments as path to a file list")
	flag.BoolVar(&local, "local", false, "use local files")
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
	if filelist {
		putFileList()
		return
	}
	ms, err := filepath.Glob(put)
	if err != nil {
		log.Fatal(err)
	}
	for _, path := range ms {
		putFileOrDir(path)
	}
}

func putFileOrDir(path string) {
	if isURL(path) {
		putFile(path)
		return
	}
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
	if isURL(path) {
		ts, err = client.PutURL(path, ls, resolvers())
	} else if local {
		ts, err = client.PutLocalFile(path, ls, resolvers())
	} else {
		is, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer is.Close()
		ts, err = client.PutContent(is, path, "text/plain", ls, resolvers())
	}
	if err != nil {
		log.Fatal(err)
	}
	printTokens(ts)
}

func putFileList() {
	file, err := os.Open(put)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		path := scanner.Text()
		if strings.HasPrefix(path, "#") || len(path) == 0 {
			continue
		}
		putFileOrDir(path)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://")
}

func resolvers() []rest.Resolver {
	var res []rest.Resolver
	for _, r := range rs {
		resolver := rest.Resolver{Name: r, MemorySize: memsize}
		if strings.ToLower(r) == "automatic" {
			resolver.Threshold = threshold
		}
		res = append(res, resolver)
	}
	return res
}

func printTokens(ts rest.Tokens) {
	sort.Slice(ts.Tokens, func(i, j int) bool {
		return ts.Tokens[i].Path < ts.Tokens[j].Path
	})
	for i, t := range ts.Tokens {
		fmt.Printf("[%d/%d] %q %q %q %q\n",
			i+1, len(ts.Tokens), t.Token, t.RelationURL,
			t.Concept.ShortName(), t.Path)
	}
}

func printConcepts(cs []*semix.Concept) {
	for i, c := range cs {
		fmt.Printf("[%d/%d] %s\n", i+1, len(cs), c)
	}
}

func assertSearchOK() {
	if id == 0 && url == "" {
		log.Fatal("missing concept id or url")
	}
}
