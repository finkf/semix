package semix

// IndexEntry denotes a public available index entry
type IndexEntry struct {
	ConceptURL, Path, OriginURL, OriginRelationURL string
	Begin, End                                     int
	Token                                          string
}

// Index represents the basic interface to put and get tokens from an index.
type Index interface {
	Put(Token) error
	Get(*Concept, func(IndexEntry)) error
	Close() error
}
