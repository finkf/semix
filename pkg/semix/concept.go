package semix

import "strings"

type Edge struct {
	P, O *Concept
}

var splitRelation = &Concept{url: "[split]"}

type Concept struct {
	url   string
	edges []Edge
	id    int32
}

func NewConcept(url string) *Concept {
	return &Concept{url: url}
}

func (c *Concept) ID() int32 {
	return c.id
}

func (c *Concept) Edges(f func(Edge)) {
	for _, e := range c.edges {
		f(e)
	}
}

func (c *Concept) URL() string {
	return c.url
}

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
