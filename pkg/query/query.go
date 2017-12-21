package query

import (
	"fmt"
	"sort"

	"bitbucket.org/fflo/semix/pkg/index"
)

// LookupFunc looks up a query string. It should return the corresponding
// URLs for the given string. It is not considered an error if the function
// returns an empty slice.
// The function should return an error if the query should fail.
type LookupFunc func(string) ([]string, error)

// Query represents a query.
type Query struct {
	constraint constraint
	set        set
	l          int
	a          bool
}

// New create a new query object from a query.
func New(query string, lookup LookupFunc) (*Query, error) {
	q, err := NewParser(query).Parse()
	if err != nil {
		return nil, err
	}
	if err := q.fix(lookup); err != nil {
		return nil, err
	}
	return q, nil
}

// fix the URLs in the constraint and query sets.
func (q *Query) fix(lookup LookupFunc) error {
	newc := make(set, len(q.constraint.set))
	for url := range q.constraint.set {
		urls, err := lookup(url)
		if err != nil {
			return err
		}
		for _, url := range urls {
			newc[url] = true
		}
	}
	news := make(set, len(q.set))
	for url := range q.set {
		urls, err := lookup(url)
		if err != nil {
			return err
		}
		for _, url := range urls {
			news[url] = true
		}
	}
	q.constraint.set = newc
	q.set = news
	return nil
}

// Execute executes the query on the given index and returns
// the slice of the matched IndexEntries.
func (q Query) Execute(idx index.Interface) ([]index.Entry, error) {
	var es []index.Entry
	err := q.ExecuteFunc(idx, func(e index.Entry) bool {
		es = append(es, e)
		return true
	})
	if err != nil {
		return nil, err
	}
	return es, nil
}

// ExecuteFunc executes the query on an index. The callback function
// is called for every matched IndexEntry.
func (q Query) ExecuteFunc(idx index.Interface, f func(index.Entry) bool) error {
	for url := range q.set {
		err := idx.Get(url, func(e index.Entry) bool {
			if q.match(e) {
				return f(e)
			}
			return true
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (q Query) match(e index.Entry) bool {
	return q.a == e.Ambiguous && e.L <= q.l && q.constraint.match(e)
}

// String returns a string representing the query.
func (q Query) String() string {
	pre := "?"
	if q.a {
		pre += "*"
	}
	if q.l != 0 {
		pre += fmt.Sprintf("%d", q.l)
	}
	c := q.constraint.String()
	if len(c) == 0 {
		return pre + "(" + q.set.String() + ")"
	}
	return pre + "(" + c + "(" + q.set.String() + "))"
}

type set map[string]bool

func (s set) String() string {
	if len(s) == 0 {
		return ""
	}
	sep := ""
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
	return str
}

func (s set) in(url string) bool {
	_, ok := s[url]
	return ok
}

type constraint struct {
	set          set
	not, all, up bool
}

func (c constraint) String() string {
	if c.all {
		return "*"
	}
	if c.not {
		return "!" + c.set.String()
	}
	return c.set.String()
}

// not & in   -> false
// !not & !in -> false
// !not & in  -> true
// not & !in  -> true
func (c constraint) match(i index.Entry) bool {
	// direct hits (with no relation URL) are always returned!
	if i.RelationURL == "" {
		return true
	}
	return c.not != c.in(i.RelationURL)
}

func (c constraint) in(url string) bool {
	if c.all {
		return url != ""
	}
	return c.set.in(url)
}
