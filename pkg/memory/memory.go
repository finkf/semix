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

// NewMemory creates a new Memory with a fixed size.
func NewMemory(n int) *Memory {
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

// Elements returns the set of unique concepts in the memory.
func (m Memory) Elements() []*semix.Concept {
	set := make(map[string]*semix.Concept)
	m.Each(func(c *semix.Concept) {
		set[c.URL()] = c
	})
	res := make([]*semix.Concept, 0, len(set))
	for _, c := range set {
		res = append(res, c)
	}
	return res
}

// Count returns the number of concepts with the given url in the memory.
func (m Memory) Count(url string) int {
	var n int
	m.Each(func(c *semix.Concept) {
		if c.URL() == url {
			n++
		}
	})
	return n
}

// N returns the size of them memory.
func (m Memory) N() int {
	return m.end
}

func inc(i uint, n int) uint {
	return (i + 1) % uint(n)
}
