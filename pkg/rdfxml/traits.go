package rdfxml

type traits struct {
	transitive, symmetric, ignore map[string]bool
	splitRelationURL              string
}

func newTraits() traits {
	return traits{
		transitive:       make(map[string]bool),
		symmetric:        make(map[string]bool),
		ignore:           make(map[string]bool),
		splitRelationURL: "http://gitlab.com/finkf/semix/pkg/rdfxml/split",
	}
}

func (t traits) isTransitiveURL(url string) bool {
	_, ok := t.transitive[url]
	return ok
}

func (t traits) isSymmetricURL(url string) bool {
	_, ok := t.symmetric[url]
	return ok
}

func (t traits) ignoreURL(url string) bool {
	_, ok := t.ignore[url]
	return ok
}
