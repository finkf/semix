package resolve

import (
	"testing"

	"bitbucket.org/fflo/semix/pkg/memory"
	"bitbucket.org/fflo/semix/pkg/semix"
)

func TestSimple(t *testing.T) {
	split := semix.NewConcept(semix.SplitURL)
	a := semix.NewConcept("A", semix.WithEdges())
	b := semix.NewConcept("B", semix.WithEdges())
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

func checkResolve(t *testing.T, got, want *semix.Concept) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %p; got %p", want, got)
	}
}
