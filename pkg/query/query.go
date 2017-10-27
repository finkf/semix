package query

import (
	"fmt"
	"sort"

	index "bitbucket.org/fflo/semix/pkg/index"
)

// Execute executes a query on the given index and returns a slice
// with all the matched IndexEntries.
func Execute(query string, idx index.Index) ([]index.Entry, error) {
	q, err := New(query)
	if err != nil {
		return nil, err
	}
	return q.Execute(idx)
}

// ExecuteFunc executes a query on the given index.
// The callback is called for every matched IndexEntry.
func ExecuteFunc(query string, idx index.Index, f func(index.Entry)) error {
	q, err := New(query)
	if err != nil {
		return err
	}
	return q.ExecuteFunc(idx, f)
}

// Query represents a query.
type Query struct {
	constraint constraint
	set        set
	l          int
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

// NewFix returns a new query and updates all urls in the query
// with the given fix function.
func NewFix(query string, fix func(string) (string, error)) (Query, error) {
	q, err := New(query)
	if err != nil {
		return Query{}, err
	}
	q1 := Query{
		constraint: constraint{
			set: make(map[string]bool),
			not: q.constraint.not,
			all: q.constraint.all,
		},
		set: make(map[string]bool),
	}
	for url := range q.constraint.set {
		newurl, err := fix(url)
		if err != nil {
			return Query{}, err
		}
		q1.constraint.set[newurl] = true
	}
	for url := range q.set {
		newurl, err := fix(url)
		if err != nil {
			return Query{}, err
		}
		q1.set[newurl] = true
	}
	return q1, nil
}

// Execute executes the query on the given index and returns
// the slice of the matched IndexEntries.
func (q Query) Execute(idx index.Index) ([]index.Entry, error) {
	var es []index.Entry
	err := q.ExecuteFunc(idx, func(e index.Entry) {
		es = append(es, e)
	})
	if err != nil {
		return nil, err
	}
	return es, nil
}

// ExecuteFunc executes the query on an index. The callback function
// is called for every matched IndexEntry.
func (q Query) ExecuteFunc(idx index.Index, f func(index.Entry)) error {
	for url := range q.set {
		err := idx.Get(url, func(e index.Entry) {
			if e.L <= q.l && q.constraint.match(e) {
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
	c := q.constraint.String()
	pre := "?"
	if q.l != 0 {
		pre += fmt.Sprintf("%d", q.l)
	}
	if len(c) == 0 {
		return pre + "({" + q.set.String() + "})"
	}
	return pre + "(" + c + "({" + q.set.String() + "}))"
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
	if url == "" {
		return len(s) == 0
	}
	_, ok := s[url]
	return ok
}

type constraint struct {
	set          set
	not, all, up bool
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
func (c constraint) match(i index.Entry) bool {
	if c.not && i.RelationURL == "" {
		return false
	}
	return c.not != c.in(i.RelationURL)
}

func (c constraint) in(url string) bool {
	if c.all {
		return url != ""
	}
	return c.set.in(url)
}
