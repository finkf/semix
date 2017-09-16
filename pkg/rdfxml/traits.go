package rdfxml

type traits struct {
	transitive, symmetric map[string]bool
	splitURL              string
}

func newTraits() traits {
	return traits{
		transitive: make(map[string]bool),
		symmetric:  make(map[string]bool),
		splitURL:   "http://bitbucket.org/fflo/semix/pkg/rdfxml/split",
	}
}
