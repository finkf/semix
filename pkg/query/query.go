package query

import "sort"

// IndexEntry is a fake struct for the basic Index interface (see below).
type IndexEntry struct {
	URL, Rel string
}

// Index represents a fake index that needs to be removed if the
// query and index branches are moved
type Index interface {
	Get(string, func(IndexEntry)) error
}

// Execute executes a query on the given index and returns a slice
// with all the matched IndexEntries.
func Execute(query string, index Index) ([]IndexEntry, error) {
	q, err := New(query)
	if err != nil {
		return nil, err
	}
	return q.Execute(index)
}

// ExecuteFunc executes a query on the given index.
// The callback is called for every matched IndexEntry.
func ExecuteFunc(query string, index Index, f func(IndexEntry)) error {
	q, err := New(query)
	if err != nil {
		return err
	}
	return q.ExecuteFunc(index, f)
}

// Query represents a query.
type Query struct {
	constraint constraint
	set        set
}

// New create a new query object from a query.
func New(query string) (Query, error) {
	p := NewParser(query)
	q, err := p.Parse()
	if err != nil {
		return Query{}, err
	}
	return q, nil
}

// Execute executes the query on the given index and returns
// the slice of the matched IndexEntries.
func (q Query) Execute(index Index) ([]IndexEntry, error) {
	var es []IndexEntry
	err := q.ExecuteFunc(index, func(e IndexEntry) {
		es = append(es, e)
	})
	if err != nil {
		return nil, err
	}
	return es, nil
}

// ExecuteFunc executes the query on an index. The callback function
// is called for every matched IndexEntry.
func (q Query) ExecuteFunc(index Index, f func(IndexEntry)) error {
	for url := range q.set {
		err := index.Get(url, func(e IndexEntry) {
			if q.constraint.match(e) {
				f(e)
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// String returns a string representing the query.
func (q Query) String() string {
	return "?(" + q.constraint.String() + "(" + q.set.String() + "))"
}

type set map[string]bool

func (s set) String() string {
	if len(s) == 0 {
		return "{}"
	}
	sep := "{"
	str := ""
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		str += sep + k
		sep = ","
	}
	return str + "}"
}

func (s set) in(url string) bool {
	if url == "" {
		return len(s) == 0
	}
	_, ok := s[url]
	return ok
}

type constraint struct {
	set      set
	not, all bool
}

func (c constraint) String() string {
	str := ""
	if c.not {
		str = "!"
	}
	if c.all {
		return str + "*"
	}
	return str + c.set.String()
}

// not & in   -> false
// !not & !in -> false
// !not & in  -> true
// not & !in  -> true
func (c constraint) match(i IndexEntry) bool {
	if c.not && i.Rel == "" {
		return false
	}
	return c.not != c.in(i.Rel)
}

func (c constraint) in(url string) bool {
	if c.all {
		return url != ""
	}
	return c.set.in(url)
}
