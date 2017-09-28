package semix

import "strings"

// Edge represents an edge in the concept graph that
// links on concept to another concept under a predicate.
type Edge struct {
	P, O *Concept
}

// TODO: do we need this?
var splitRelation = &Concept{url: "[split]"}

// Concept represents a concept in the concept graph.
// It consits of an unique URL, an optional (human readeable) name,
// a list of edges and an unique ID.
// TODO: do we need the ID?
type Concept struct {
	url, name string
	edges     []Edge
	id        int32
}

// NewConcept create a new Concept with the given URL.
func NewConcept(url string) *Concept {
	return &Concept{url: url}
}

// ID returns the unique ID of the concept.
func (c *Concept) ID() int32 {
	return c.id
}

// EdgesLen returns the length of the edges.
func (c *Concept) EdgesLen() int {
	return len(c.edges)
}

// EdgeAt returns the edge at the given position in the edges slice.
func (c *Concept) EdgeAt(i int) Edge {
	return c.edges[i]
}

// Edges iterates over all edges of this. concept.
func (c *Concept) Edges(f func(Edge)) {
	for _, e := range c.edges {
		f(e)
	}
}

// URL return the url of this concept.
func (c *Concept) URL() string {
	return c.url
}

// ShortURL returns a short version of the URL of this concept.
// The short URL is not necessarily unique.
func (c *Concept) ShortURL() string {
	i := strings.LastIndex(c.url, "/")
	if i == -1 {
		return c.url
	}
	return c.url[i+1:]
}

func (c *Concept) String() string {
	return c.ShortURL()
}
