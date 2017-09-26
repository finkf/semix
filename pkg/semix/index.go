package semix

// IndexEntry denotes a public available index entry
type IndexEntry struct {
	ConceptURL, Path, OriginURL, RelationURL string
	Begin, End                               int
	Token                                    string
}

// Index represents the basic interface to put and get tokens from an index.
type Index interface {
	Put(Token) error
	Get(string, func(IndexEntry)) error
	Close() error
}
