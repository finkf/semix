package traits

import "bitbucket.org/fflo/semix/pkg/semix"

// Option speciefies an Option to set up a traits instance.
type Option func(traits)

// WithIgnoreURLs specifies a list of symmetric predicate URLs.
func WithIgnoreURLs(strs ...string) Option {
	return func(t traits) {
		for _, str := range strs {
			t.i[str] = true
		}
	}
}

// WithTransitiveURLs specifies a list of transitive predicate URLs.
func WithTransitiveURLs(urls ...string) Option {
	return func(t traits) {
		for _, url := range urls {
			t.t[url] = true
		}
	}
}

// WithSymmetricURLs specifies a list of symmetric predicate URLs.
func WithSymmetricURLs(urls ...string) Option {
	return func(t traits) {
		for _, url := range urls {
			t.s[url] = true
		}
	}
}

// WithNameURLs specifies a list of name predicate URLs.
func WithNameURLs(urls ...string) Option {
	return func(t traits) {
		for _, url := range urls {
			t.n[url] = true
		}
	}
}

// WithDistinctURLs specifies a list of distinct predicate URLs.
func WithDistinctURLs(urls ...string) Option {
	return func(t traits) {
		for _, url := range urls {
			t.d[url] = true
		}
	}
}

// WithAmbiguousURLs specifies a list of ambiguous predicate URLs.
func WithAmbiguousURLs(urls ...string) Option {
	return func(t traits) {
		for _, url := range urls {
			t.a[url] = true
		}
	}
}

// New returns a new traits instance.
// You can set the various urls manually.
func New(opts ...Option) semix.Traits {
	t := traits{
		i: make(traitSet),
		t: make(traitSet),
		s: make(traitSet),
		n: make(traitSet),
		d: make(traitSet),
		a: make(traitSet),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

type traitSet map[string]bool

type traits struct {
	i, t, s, n, d, a traitSet
}

func (t traits) Ignore(url string) bool {
	return t.i[url]
}

func (t traits) IsSymmetric(url string) bool {
	return t.s[url]
}

func (t traits) IsTransitive(url string) bool {
	return t.t[url]
}

func (t traits) IsName(url string) bool {
	return t.n[url]
}

func (t traits) IsDistinct(url string) bool {
	return t.d[url]
}

func (t traits) IsAmbiguous(url string) bool {
	return t.d[url]
}
