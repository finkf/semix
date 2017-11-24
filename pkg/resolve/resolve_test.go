package resolve

import (
	"testing"

	"bitbucket.org/fflo/semix/pkg/memory"
	"bitbucket.org/fflo/semix/pkg/semix"
)

func TestSimple(t *testing.T) {
	split := semix.NewConcept(semix.SplitURL)
	a := semix.NewConcept("A")
	b := semix.NewConcept("B")
	ambig := semix.NewConcept("A-B", semix.WithEdges(split, a, split, b))
	mem := memory.New(3)
	var simple Simple
	checkResolve(t, simple.Resolve(ambig, mem), nil)
	mem.Push(a)
	checkResolve(t, simple.Resolve(ambig, mem), a)
	mem.Push(b)
	checkResolve(t, simple.Resolve(ambig, mem), nil)
	mem.Push(b)
	checkResolve(t, simple.Resolve(ambig, mem), b)
	mem.Push(a)
	checkResolve(t, simple.Resolve(ambig, mem), b)
	mem.Push(a)
	checkResolve(t, simple.Resolve(ambig, mem), a)
}

func TestAutomatic(t *testing.T) {
	split := semix.NewConcept(semix.SplitURL)
	br := semix.NewConcept("broader")
	p := semix.NewConcept("politics")
	q := semix.NewConcept("quantum-physics")
	a := semix.NewConcept("A", semix.WithEdges(br, p))
	b := semix.NewConcept("B", semix.WithEdges(br, q))
	ambig := semix.NewConcept("A-B", semix.WithEdges(split, a, split, b))
	mem := memory.New(3)
	automatic := Automatic{0.5}
	checkResolve(t, automatic.Resolve(ambig, mem), nil)
	mem.Push(p)
	checkResolve(t, automatic.Resolve(ambig, mem), a)
	mem.Push(a)
	// overlap = 0.5
	checkResolve(t, automatic.Resolve(ambig, mem), nil)
	mem.Push(q)
	checkResolve(t, automatic.Resolve(ambig, mem), nil)
	mem.Push(q)
	checkResolve(t, automatic.Resolve(ambig, mem), nil)
	mem.Push(q)
	checkResolve(t, automatic.Resolve(ambig, mem), b)
}

func checkResolve(t *testing.T, got, want *semix.Concept) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %p; got %p", want, got)
	}
}
