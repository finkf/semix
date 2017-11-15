package disamb

import (
	"bitbucket.org/fflo/semix/pkg/semix"
)

// Memory is simple ringbuffer that enables counting of concepts stored
// in the memory.
type Memory struct {
	buffer []*semix.Concept
	i, end uint
}

// NewMemory creates a new Memory with a fixed size.
func NewMemory(n int) *Memory {
	return &Memory{buffer: make([]*semix.Concept, n)}
}

func inc(i uint, n int) uint {
	return (i + 1) % uint(n)
}
