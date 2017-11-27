package memory

import (
	"bitbucket.org/fflo/semix/pkg/semix"
)

// Memory is simple ringbuffer that enables counting of concepts stored
// in the memory.
type Memory struct {
	buffer []*semix.Concept
	i      uint
	end    int
}

// New creates a new Memory with a fixed size.
func New(n int) *Memory {
	return &Memory{buffer: make([]*semix.Concept, n)}
}

// Push pushes a new concept into the memory.
// It removes the last inserted concept.
func (m *Memory) Push(c *semix.Concept) {
	if m.end < len(m.buffer) {
		m.end++
	}
	m.buffer[m.i] = c
	m.i = inc(m.i, len(m.buffer))
}

// Each calls a callback function for each concept in the memory.
func (m Memory) Each(f func(*semix.Concept)) {
	for i := 0; i < m.end; i++ {
		f(m.buffer[i])
	}
}

// EachS calls a callback function for each concept in the memory and for
// each concept that is referenced from this concept.
func (m Memory) EachS(f func(*semix.Concept)) {
	for i := 0; i < m.end; i++ {
		f(m.buffer[i])
		m.buffer[i].EachEdge(func(e semix.Edge) {
			f(e.O)
		})
	}
}

// Elements returns the set of unique concepts in the memory.
func (m Memory) Elements() map[string]*semix.Concept {
	set := make(map[string]*semix.Concept)
	m.Each(func(c *semix.Concept) {
		set[c.URL()] = c
	})
	return set
}

// ElementsS returns the set of unique concepts in the memory,
// including all referenced concepts.
func (m Memory) ElementsS() map[string]*semix.Concept {
	set := make(map[string]*semix.Concept)
	m.EachS(func(c *semix.Concept) {
		set[c.URL()] = c
	})
	return set
}

// CountIf returns the number of concepts for wich the given predicate returns true.
func (m Memory) CountIf(p func(c *semix.Concept) bool) int {
	var n int
	m.Each(func(c *semix.Concept) {
		if p(c) {
			n++
		}
	})
	return n
}

// CountIfS returns the number of concepts for wich the given predicate returns true.
// The predicate is called for each concept and referenced concept.
func (m Memory) CountIfS(p func(c *semix.Concept) bool) int {
	var n int
	m.EachS(func(c *semix.Concept) {
		if p(c) {
			n++
		}
	})
	return n
}

// N returns the maximal number of elements in the memory.
func (m Memory) N() int {
	return len(m.buffer)
}

// Len returns the number of elements in the memory.
func (m Memory) Len() int {
	return m.end
}

func inc(i uint, n int) uint {
	return (i + 1) % uint(n)
}

func sliceFromSet(set map[string]*semix.Concept) []*semix.Concept {
	res := make([]*semix.Concept, 0, len(set))
	for _, c := range set {
		res = append(res, c)
	}
	return res
}
