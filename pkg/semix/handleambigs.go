package semix

import "sort"

// HandleAmbiguitiesFunc defines a function that handles ambiguities
// in the parsing of the knowledge base.
// If the function is successfull, it must return a non nil concept,
// otherwise the according dictionary entry is discarded.
type HandleAmbiguitiesFunc func(*Graph, string, ...string) *Concept

// HandleAmbiguitiesWithSplit handles an ambiguity
// by creating a new ambig split concept.
func HandleAmbiguitiesWithSplit(g *Graph, entry string, urls ...string) *Concept {
	urls = sortUnique(urls)
	newURL := CombineURLs(urls...)
	c := g.Register(newURL)
	for _, url := range urls {
		g.Add(newURL, SplitURL, url)
	}
	return c
}

// HandleAmbiguitiesWithMerge handles ambiguities
// by creating a new distinct concept.
func HandleAmbiguitiesWithMerge(g *Graph, entry string, urls ...string) *Concept {
	urls = sortUnique(urls)
	newURL := CombineURLs(urls...)
	edges := intersectEdges(g, urls...)
	c := g.Register(newURL)
	for p, os := range edges {
		for o := range os {
			g.Add(newURL, p, o)
		}
	}
	return c

}

func intersectEdges(g *Graph, urls ...string) map[string]map[string]struct{} {
	if len(urls) == 0 {
		return nil
	}
	a := makeEdgesMap(g, urls[0])
	for _, url := range urls[1:] {
		a = intersect(a, makeEdgesMap(g, url))
	}
	return a
}

func intersect(a, b map[string]map[string]struct{}) map[string]map[string]struct{} {
	c := make(map[string]map[string]struct{})
	for p, os := range a {
		for o := range os {
			if _, ok := b[p]; ok {
				if _, ok := b[p][o]; ok {
					if _, ok := c[p]; !ok {
						c[p] = make(map[string]struct{})
					}
					c[p][o] = struct{}{}
				}
			}
		}
	}
	return c
}

func makeEdgesMap(g *Graph, url string) map[string]map[string]struct{} {
	c, ok := g.FindByURL(url)
	if !ok {
		return nil
	}
	edges := make(map[string]map[string]struct{})
	for _, e := range c.edges {
		if edges[e.P.URL()] == nil {
			edges[e.P.URL()] = make(map[string]struct{})
		}
		edges[e.P.URL()][e.O.URL()] = struct{}{}
	}
	return edges
}

func sortUnique(urls []string) []string {
	urlset := make(map[string]bool)
	for _, url := range urls {
		urlset[url] = true
	}
	splits := make([]string, 0, len(urlset))
	for url := range urlset {
		splits = append(splits, url)
	}
	sort.Strings(splits)
	return splits
}
