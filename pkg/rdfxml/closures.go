package rdfxml

func calculateTransitiveClosure(ts map[triple]bool) map[triple]bool {
	visited := make(map[triple]bool, len(ts))
	closure := make(map[triple]bool, len(ts)*2)
	flat := flatten(ts)
	out := outgoing(ts)
	for i := 0; i < len(flat); i++ {
		current := flat[i]
		if _, ok := visited[current]; ok {
			continue
		}
		visited[current] = true
		closure[current] = true
		for _, t := range out[current.o] {
			newtriple := triple{s: current.s, p: current.p, o: t}
			closure[newtriple] = true
			flat = append(flat, newtriple)
		}
	}
	return closure
}

func outgoing(ts map[triple]bool) map[string][]string {
	out := make(map[string][]string, len(ts))
	for t := range ts {
		out[t.s] = append(out[t.s], t.o)
	}
	return out
}

func flatten(ts map[triple]bool) []triple {
	flat := make([]triple, 0, len(ts))
	for t := range ts {
		flat = append(flat, t)
	}
	return flat
}

func calculateSymmetricClosure(ts map[triple]bool) map[triple]bool {
	closure := make(map[triple]bool, len(ts))
	for t := range ts {
		closure[t] = true
		closure[triple{s: t.o, p: t.p, o: t.s}] = true
	}
	return closure
}
