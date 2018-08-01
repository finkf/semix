package resolve

import (
	"context"
	"testing"

	"gitlab.com/finkf/semix/pkg/memory"
	"gitlab.com/finkf/semix/pkg/rule"
	"gitlab.com/finkf/semix/pkg/semix"
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

func TestRuled(t *testing.T) {
	split := semix.NewConcept(semix.SplitURL)
	a := semix.NewConcept("A", semix.WithID(1))
	b := semix.NewConcept("B", semix.WithID(2))
	ambig := semix.NewConcept("A-B", semix.WithEdges(split, a, split, b))
	rules, err := rule.NewMap(map[string]string{"A": `cs("A")>0`, "B": `cs("B")>0`}, func(str string) int {
		switch str {
		case "A":
			return 1
		case "B":
			return 2
		default:
			return -1
		}
	})
	if err != nil {
		t.Fatalf("got error: %s", err)
	}
	ruled := Ruled{Rules: rules}
	mem := memory.New(3)
	checkResolve(t, ruled.Resolve(ambig, mem), nil)
	mem.Push(a)
	checkResolve(t, ruled.Resolve(ambig, mem), a)
	mem.Push(b)
	checkResolve(t, ruled.Resolve(ambig, mem), nil)
	mem.Push(b)
	checkResolve(t, ruled.Resolve(ambig, mem), nil)
	mem.Push(b)
	checkResolve(t, ruled.Resolve(ambig, mem), b)
}

func TestStream(t *testing.T) {
	split := semix.NewConcept(semix.SplitURL)
	a := semix.NewConcept("A")
	b := semix.NewConcept("B")
	ambig := semix.NewConcept("A-B", semix.WithEdges(split, a, split, b))
	var simple Simple
	tokens := make(chan semix.StreamToken)
	go func() {
		tokens <- semix.StreamToken{Token: semix.Token{Concept: a, Path: "test"}}
		tokens <- semix.StreamToken{Token: semix.Token{Concept: ambig, Path: "test"}}
		tokens <- semix.StreamToken{Token: semix.Token{Concept: b, Path: "test"}}
		close(tokens)
	}()
	counts := make(map[string]int)
	for tok := range Resolve(context.TODO(), 3, simple, tokens) {
		if tok.Err != nil {
			t.Fatalf("go error: %s", tok.Err)
		}
		counts[tok.Token.Concept.URL()]++
	}
	if got := counts["A"]; got != 2 {
		t.Fatalf("expected %d; got %d", 2, got)
	}
	if got := counts["B"]; got != 1 {
		t.Fatalf("expected %d; got %d", 1, got)
	}
	if got := len(counts); got != 2 {
		t.Fatalf("expected %d; got %d", 2, got)
	}
}

func checkResolve(t *testing.T, got, want *semix.Concept) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %p; got %p", want, got)
	}
}
