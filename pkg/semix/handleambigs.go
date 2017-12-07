package semix

import "sort"

// HandleAmbigsFunc defines a function that handles ambiguities
// in the parsing of the knowledge base.
// If the function is successfull, it must return a non nil concept,
// otherwise the according dictionary entry is discarded.
type HandleAmbigsFunc func(*Graph, string, ...string) *Concept

// HandleAmbigsWithSplit handles an ambiguity
// by creating a new ambig split concept.
func HandleAmbigsWithSplit(g *Graph, entry string, urls ...string) *Concept {
	urls = sortUnique(urls)
	newURL := CombineURLs(urls...)
	c := g.Register(newURL)
	for _, url := range urls {
		g.Add(newURL, SplitURL, url)
	}
	return c
}

// HandleAmbigsWithMerge handles ambiguities
// by creating a new distinct concept.
func HandleAmbigsWithMerge(g *Graph, entry string, urls ...string) *Concept {
	urls = sortUnique(urls)
	newURL := CombineURLs(urls...)
	edges := IntersectEdges(g, urls...)
	c := g.Register(newURL)
	for p, os := range edges {
		for o := range os {
			g.Add(newURL, p, o)
		}
	}
	return c
}

// EdgeSet represents a set of relations
type EdgeSet map[string]map[string]struct{}

// IntersectEdges calculates the intersection of the
// relation sets of the given concepts.
func IntersectEdges(g *Graph, urls ...string) EdgeSet {
	if len(urls) == 0 {
		return nil
	}
	a := newEdgeSet(g, urls[0])
	for _, url := range urls[1:] {
		a = intersect(a, newEdgeSet(g, url))
	}
	return a
}

func intersect(a, b EdgeSet) EdgeSet {
	c := make(EdgeSet)
	for p, os := range a {
		for o := range os {
			if _, ok := b[p]; !ok {
				continue
			}
			if _, ok := b[p][o]; !ok {
				continue
			}
			if _, ok := c[p]; !ok {
				c[p] = make(map[string]struct{})
			}
			c[p][o] = struct{}{}
		}
	}
	return c
}

func newEdgeSet(g *Graph, url string) map[string]map[string]struct{} {
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
