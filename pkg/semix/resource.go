package semix

import (
	"bytes"
	"encoding/gob"
)

// Dictionary is a dictionary that maps the labels of the concepts
// to their apporpriate IDs. Negative IDs mark ambigous dictionary entries.
// The map to the according positve ID.
type Dictionary map[string]int32

// RulesDictionary is a dictionary that maps concept URLs to their
// respective rules.
type RulesDictionary map[string]string

// Resource is a struct that holds all parsed knwoledge base resources.
type Resource struct {
	Graph      *Graph
	Dictionary Dictionary
	Rules      RulesDictionary
	DFA        DFA
}

// NewResource creates a new resource.
func NewResource(g *Graph, d Dictionary, r RulesDictionary) *Resource {
	return &Resource{
		Graph:      g,
		Dictionary: d,
		Rules:      r,
		DFA:        NewDFA(d, g),
	}
}

type gobEdge struct {
	P, O int32
}

// GobDecode decodes a graph from gob endcoded binary data.
func (r *Resource) GobDecode(bs []byte) error {
	decoder := gob.NewDecoder(bytes.NewBuffer(bs))
	if err := decoder.Decode(&r.DFA); err != nil {
		return err
	}
	if err := decoder.Decode(&r.Dictionary); err != nil {
		return err
	}
	if err := decoder.Decode(&r.Rules); err != nil {
		return err
	}
	r.Graph = new(Graph)
	if err := decoder.Decode(&r.Graph.cArr); err != nil {
		return err
	}
	register := make(map[int32][]gobEdge)
	if err := decoder.Decode(&register); err != nil {
		return err
	}
	r.Graph.cMap = make(map[string]*Concept)
	for _, c := range r.Graph.cArr {
		r.Graph.cMap[c.url] = c
		for _, e := range register[c.id] {
			c.edges = append(c.edges, Edge{
				P: r.Graph.cArr[e.P-1],
				O: r.Graph.cArr[e.O-1],
			})
		}
	}
	r.DFA.graph = r.Graph
	return nil
}

// GobEncode encodes a graph to gob encoded binary data.
func (r *Resource) GobEncode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	register := make(map[int32][]gobEdge)
	for _, c := range r.Graph.cArr {
		for _, e := range c.edges {
			register[c.id] = append(register[c.id], gobEdge{e.P.id, e.O.id})
		}
	}
	if err := encoder.Encode(r.DFA); err != nil {
		return nil, err
	}
	if err := encoder.Encode(r.Dictionary); err != nil {
		return nil, err
	}
	if err := encoder.Encode(r.Rules); err != nil {
		return nil, err
	}
	if err := encoder.Encode(r.Graph.cArr); err != nil {
		return nil, err
	}
	if err := encoder.Encode(register); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// GobDecode decodes a concept from gob encoded binary data.
// Only the name, url and id are decoded.
func (c *Concept) GobDecode(bs []byte) error {
	decoder := gob.NewDecoder(bytes.NewBuffer(bs))
	var url, name string
	var id int32
	if err := decoder.Decode(&url); err != nil {
		return err
	}
	if err := decoder.Decode(&name); err != nil {
		return err
	}
	if err := decoder.Decode(&id); err != nil {
		return err
	}
	c.url = url
	c.Name = name
	c.id = id
	return nil
}

// GobEncode encodes a concept to gob encoded binary data.
// Only the name, url and id are encoded.
func (c *Concept) GobEncode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	if err := encoder.Encode(c.url); err != nil {
		return nil, err
	}
	if err := encoder.Encode(c.Name); err != nil {
		return nil, err
	}
	if err := encoder.Encode(c.id); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// GobDecode decodes a the sparsetable.DFA of a DFA.
// It does not decode the graph.
func (d *DFA) GobDecode(bs []byte) error {
	decoder := gob.NewDecoder(bytes.NewBuffer(bs))
	return decoder.Decode(&d.dfa)
}

// GobEncode encodes a the sparsetable.DFA of a DFA.
// It does not encode the graph.
func (d DFA) GobEncode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	if err := encoder.Encode(d.dfa); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
