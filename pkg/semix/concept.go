package semix

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const (
	// SplitURL is the name of predicates that denote ambiguous connections
	// in the concept graph.
	SplitURL = "http://bitbucket.org/fflo/semix/pkg/semix/a-star"
)

// Edge represents an edge in the concept graph that
// links on concept to another concept with a predicate and a Levenshtein distance.
type Edge struct {
	P, O *Concept
	L    int
}

func (e Edge) String() string {
	return fmt.Sprintf("{%s %s %d}", e.P.ShortName(), e.O.ShortName(), e.L)
}

// Concept represents a concept in the concept graph.
// It consits of an unique URL, an optional (human readeable) name,
// a list of edges and an unique ID.
type Concept struct {
	url, Name string
	edges     []Edge
	id        int32
}

// Edges returns the edges of a concept.
func (c Concept) Edges() []Edge {
	return c.edges
}

// HasLink returns true if the concept has an outgoing
// edge to o.
func (c Concept) HasLink(o *Concept) bool {
	for _, e := range c.edges {
		if e.O.ID() == o.ID() {
			return true
		}
	}
	return false
}

// HasLinkP returns true if the concept has an outgoing
// edge to o with predicate p.
func (c Concept) HasLinkP(p, o *Concept) bool {
	for _, e := range c.edges {
		if e.P.ID() == p.ID() && o.ID() == e.O.ID() {
			return true
		}
	}
	return false
}

// CombineURLs combines tow or more URLs.
// If urls is empty, the empty string is returned.
// If urls contain exactly on url, this url is returned.
func CombineURLs(urls ...string) string {
	if len(urls) == 0 {
		return ""
	}
	if len(urls) == 1 {
		return urls[0]
	}
	res := urls[0]
	for i := 1; i < len(urls); i++ {
		res = combineTwoURLs(res, urls[i])
	}
	return res
}

func combineTwoURLs(a, b string) string {
	ai := strings.LastIndex(a, "/")
	bi := strings.LastIndex(b, "/")
	if ai == -1 || bi == -1 {
		return a + "-" + b
	}
	return a + "-" + url.PathEscape(b[bi+1:])
}

// WithID returns a configuration function that sets the concept's ID.
func WithID(id int) func(*Concept) {
	return func(c *Concept) {
		c.id = int32(id)
	}
}

// WithEdges returns a configuration function that sets the edges.
// Each pair (p,o) in cs is set to the edge pointing to o with the predicate p.
func WithEdges(cs ...*Concept) func(*Concept) {
	return func(c *Concept) {
		for i := 1; i < len(cs); i += 2 {
			c.edges = append(c.edges, Edge{P: cs[i-1], O: cs[i]})
		}
	}
}

// NewConcept creates a new Concept with the given URL
// and configuration functions.
func NewConcept(url string, cfs ...func(*Concept)) *Concept {
	c := &Concept{url: url}
	for _, cf := range cfs {
		cf(c)
	}
	return c
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

// EachEdge iterates over all edges of this concept.
func (c *Concept) EachEdge(f func(Edge)) {
	for _, e := range c.edges {
		f(e)
	}
}

// VisitAll visits all concepts (including this) recursively.
// It is guaranteed that each concept is visited exactly once.
// Predicates are ignored.
func (c *Concept) VisitAll(f func(c *Concept)) {
	visited := make(map[*Concept]bool)
	c.visit(f, visited)
}

func (c *Concept) visit(f func(c *Concept), visited map[*Concept]bool) {
	if visited[c] {
		return
	}
	visited[c] = true
	f(c)
	for _, e := range c.edges {
		e.O.visit(f, visited)
	}
}

// URL return the url of this concept.
func (c *Concept) URL() string {
	return c.url
}

// Ambig returns if the concept is ambiguous or not.
func (c *Concept) Ambig() bool {
	if len(c.edges) == 0 {
		return false
	}
	return c.edges[0].P.url == SplitURL
}

// FindEdge searches for the matching edge.
func (c *Concept) FindEdge(p, o string) (Edge, bool) {
	for _, e := range c.edges {
		if e.P.URL() == p && e.O.URL() == o {
			return e, true
		}
	}
	return Edge{}, false
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
	return fmt.Sprintf("{%s %d %v}", c.ShortName(), c.id, c.edges)
}

// ShortName returns a nice human readeable name for the concept.
// This does not need to be a unique identifier for this concept.
func (c *Concept) ShortName() string {
	if c.Name != "" {
		return c.Name
	}
	return c.ShortURL()
}

// ReduceTransitive removes all transitive edges of this
// and all its linked concepts.
func (c *Concept) ReduceTransitive() {
	len := len(c.edges)
	for i := 0; i < len; {
		e := c.edges[i]
		if c.anyEdgeHasLinkToP(e.P, e.O) {
			// swap edges
			tmp := c.edges[len-1]
			c.edges[len-1] = c.edges[i]
			c.edges[i] = tmp
			len--
		} else {
			i++
		}
	}
	for _, e := range c.edges {
		e.O.ReduceTransitive()
	}
	c.edges = c.edges[0:len]
}

func (c *Concept) anyEdgeHasLinkToP(p, o *Concept) bool {
	for _, e := range c.edges {
		if e.O.ID() != o.ID() && e.P.ID() == p.ID() && e.O.HasLinkP(p, o) {
			return true
		}
	}
	return false
}

// link represents an edge as a pair of URLs.
type link struct {
	P struct {
		URL, Name string
		ID        int
	}
	O struct {
		URL, Name string
		ID        int
	}
	L int
}

// links returns the edges of this concept as pair of URLs.
func (c *Concept) links() []link {
	links := make([]link, len(c.edges))
	for i := range c.edges {
		links[i].P.URL = c.edges[i].P.url
		links[i].P.Name = c.edges[i].P.Name
		links[i].P.ID = int(c.edges[i].P.id)
		links[i].O.URL = c.edges[i].O.url
		links[i].O.Name = c.edges[i].O.Name
		links[i].O.ID = int(c.edges[i].O.id)
		links[i].L = c.edges[i].L
	}
	return links
}

// UnmarshalJSON reads the concept from json.
// Since the edges are written as pairs of URLs,
// it is not possible to recreate the whole concept using json.
func (c *Concept) UnmarshalJSON(b []byte) error {
	var data struct {
		URL, Name string
		ID        int
		Edges     []link
		Ambiguous bool
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	*c = Concept{
		Name:  data.Name,
		url:   data.URL,
		id:    int32(data.ID),
		edges: make([]Edge, len(data.Edges)),
	}
	// create unique local concepts that users can
	// use the *Concepts as valid map entries etc.
	urls := make(map[string]*Concept)
	urls[c.url] = c // handle self references.
	for i, edge := range data.Edges {
		if _, ok := urls[edge.P.URL]; !ok {
			urls[edge.P.URL] = &Concept{
				url:  edge.P.URL,
				id:   int32(edge.P.ID),
				Name: edge.P.Name,
			}
		}
		if _, ok := urls[edge.O.URL]; !ok {
			urls[edge.O.URL] = &Concept{
				url:  edge.O.URL,
				id:   int32(edge.O.ID),
				Name: edge.O.Name,
			}
		}
		c.edges[i] = Edge{P: urls[edge.P.URL], O: urls[edge.O.URL]}
	}
	return nil
}

// MarshalJSON writes the concept to json.
// To avoid writting the whole graph of the concepts,
// the edges of the concept are written as pairs of URLs
// and recursive links are omitted.
func (c *Concept) MarshalJSON() ([]byte, error) {
	data := struct {
		URL, Name string
		ID        int
		Edges     []link
		Ambiguous bool
	}{
		URL:   c.url,
		Name:  c.Name,
		ID:    int(c.id),
		Edges: c.links(),
	}
	return json.Marshal(data)
}
