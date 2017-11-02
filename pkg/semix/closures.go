package semix

type spo struct {
	s, p, o string
}

func calculateTransitiveClosure(ts map[spo]bool) map[spo]bool {
	visited := make(map[spo]bool, len(ts))
	closure := make(map[spo]bool, len(ts)*2)
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
			newspo := spo{s: current.s, p: current.p, o: t}
			closure[newspo] = true
			flat = append(flat, newspo)
		}
	}
	return closure
}

func outgoing(ts map[spo]bool) map[string][]string {
	out := make(map[string][]string, len(ts))
	for t := range ts {
		out[t.s] = append(out[t.s], t.o)
	}
	return out
}

func flatten(ts map[spo]bool) []spo {
	flat := make([]spo, 0, len(ts))
	for t := range ts {
		flat = append(flat, t)
	}
	return flat
}

func calculateSymmetricClosure(ts map[spo]bool) map[spo]bool {
	closure := make(map[spo]bool, len(ts))
	for t := range ts {
		closure[t] = true
		closure[spo{s: t.o, p: t.p, o: t.s}] = true
	}
	return closure
}
