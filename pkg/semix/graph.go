package semix

type Triple struct {
	S, P, O *Concept
}

type Graph struct {
	cMap map[string]*Concept
	cArr []*Concept
}

func NewGraph() *Graph {
	return &Graph{
		cMap: make(map[string]*Concept),
		cArr: nil,
	}
}

func (g *Graph) FindByURL(str string) *Concept {
	if c, ok := g.cMap[str]; ok {
		return c
	}
	return nil
}

func (g *Graph) FindById(id int32) *Concept {
	if id == 0 {
		return nil
	}
	if id < 0 {
		id = -id
	}
	if int(id) > len(g.cArr) {
		return nil
	}
	return g.cArr[id-1]
}

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
