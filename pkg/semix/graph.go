package semix

// Triple represents a relational triple in the graph.
// It consitst of a subject S, a predicate P and an object O.
type Triple struct {
	S, P, O *Concept
}

// Graph represents a graph of linked concepts.
// It holds a map of the URLs and the concepts and
// an array of all concepts.
type Graph struct {
	cMap map[string]*Concept
	cArr []*Concept
}

// NewGraph creates a new graph.
func NewGraph() *Graph {
	return &Graph{
		cMap: make(map[string]*Concept),
		cArr: nil,
	}
}

// FindByURL searches a concept by its URL.
func (g *Graph) FindByURL(str string) (*Concept, bool) {
	if c, ok := g.cMap[str]; ok {
		return c, true
	}
	return nil, false
}

// FindById searches a concept by its ID.
func (g *Graph) FindById(id int32) (*Concept, bool) {
	if id == 0 {
		return nil, false
	}
	if id < 0 {
		id = -id
	}
	if int(id) > len(g.cArr) {
		return nil, false
	}
	return g.cArr[id-1], true
}

// Add adds a triple to the graph.
// It returns a Triple that consits of the according concepts
// that where created.
func (g *Graph) Add(s, p, o string) Triple {
	if s == "" || p == "" || o == "" {
		panic("cannot insert empty concept")
	}
	var triple Triple
	triple.S = g.register(s)
	triple.P = g.register(p)
	triple.O = g.register(o)
	triple.S.edges = append(triple.S.edges, Edge{P: triple.P, O: triple.O})
	return triple
}

func (g *Graph) register(url string) *Concept {
	if c, ok := g.cMap[url]; ok {
		return c
	}
	c := &Concept{url: url}
	g.cArr = append(g.cArr, c)
	c.id = int32(len(g.cArr))
	g.cMap[c.url] = c
	return c
}
