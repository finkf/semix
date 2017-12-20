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

// FindByID searches a concept by its ID.
// If a negative ID is given, a new split concept is returned,
// that links to the concept with the according positive ID.
func (g *Graph) FindByID(id int32) (*Concept, bool) {
	if id == 0 {
		return nil, false
	}
	var ambiguous bool
	if id < 0 {
		id = -id
		ambiguous = true
	}
	if int(id) > len(g.cArr) {
		return nil, false
	}
	c := g.cArr[id-1]
	if ambiguous {
		return g.makeSplitConcept(c), true
	}
	return c, true
}

// ConceptsLen returns the number of concepts in the array.
func (g *Graph) ConceptsLen() int {
	return len(g.cArr)
}

// ConceptAt returns the concept at the given position.
func (g *Graph) ConceptAt(i int) *Concept {
	return g.cArr[i]
}

// Add adds a triple to the graph.
// It returns a Triple that consits of the according concepts
// that where created.
func (g *Graph) Add(s, p, o string) (*Concept, *Concept, *Concept) {
	if s == "" || p == "" || o == "" {
		panic("cannot insert empty concept")
	}
	sc := g.Register(s)
	pc := g.Register(p)
	oc := g.Register(o)
	if _, ok := sc.FindEdge(pc.URL(), oc.URL()); !ok {
		sc.edges = append(sc.edges, Edge{P: pc, O: oc})
	}
	return sc, pc, oc
}

// Register registers new concept with the given URL in the Graph.
// If the URL does already exist, the according cocnept is retuned.
// This function will never return a nil concept.
func (g *Graph) Register(url string) *Concept {
	if c, ok := g.cMap[url]; ok {
		return c
	}
	c := &Concept{url: url}
	g.cArr = append(g.cArr, c)
	c.id = int32(len(g.cArr))
	g.cMap[c.url] = c
	return c
}

func (g *Graph) makeSplitConcept(c *Concept) *Concept {
	split, ok := g.FindByURL(SplitURL)
	if !ok {
		split = g.Register(SplitURL)
	}
	return &Concept{
		url:   c.url,
		Name:  c.Name,
		id:    c.id,
		edges: []Edge{Edge{P: split, O: c, L: -1}},
	}
}
