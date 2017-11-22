package traits

// Interface defines the interface for the different traits of predicates.
type Interface interface {
	Ignore(string) bool
	IsSymmetric(string) bool
	IsTransitive(string) bool
	IsName(string) bool
	IsDistinct(string) bool
	IsAmbiguous(string) bool
	IsInverted(string) bool
	IsRule(string) bool
}

// Option speciefies an Option to set up a traits instance.
type Option func(traits)

// WithIgnorePredicates specifies a list of symmetric predicate URLs.
func WithIgnorePredicates(strs ...string) Option {
	return func(t traits) {
		for _, str := range strs {
			t.i[str] = true
		}
	}
}

// WithTransitivePredicates specifies a list of transitive predicate URLs.
func WithTransitivePredicates(urls ...string) Option {
	return func(t traits) {
		for _, url := range urls {
			t.t[url] = true
		}
	}
}

// WithSymmetricPredicates specifies a list of symmetric predicate URLs.
func WithSymmetricPredicates(urls ...string) Option {
	return func(t traits) {
		for _, url := range urls {
			t.s[url] = true
		}
	}
}

// WithInvertedPredicates specifies a list of symmetric predicate URLs.
func WithInvertedPredicates(urls ...string) Option {
	return func(t traits) {
		for _, url := range urls {
			t.v[url] = true
		}
	}
}

// WithNamePredicates specifies a list of name predicate URLs.
func WithNamePredicates(urls ...string) Option {
	return func(t traits) {
		for _, url := range urls {
			t.n[url] = true
		}
	}
}

// WithDistinctPredicates specifies a list of distinct predicate URLs.
func WithDistinctPredicates(urls ...string) Option {
	return func(t traits) {
		for _, url := range urls {
			t.d[url] = true
		}
	}
}

// WithAmbiguousPredicates specifies a list of ambiguous predicate URLs.
func WithAmbiguousPredicates(urls ...string) Option {
	return func(t traits) {
		for _, url := range urls {
			t.a[url] = true
		}
	}
}

// WithRulesPredicates specifies a list of ambiguous predicate URLs.
func WithRulePredicates(urls ...string) Option {
	return func(t traits) {
		for _, url := range urls {
			t.r[url] = true
		}
	}
}

// New returns a new traits instance.
// You can set the various urls manually.
func New(opts ...Option) Interface {
	t := traits{
		i: make(traitSet),
		t: make(traitSet),
		s: make(traitSet),
		n: make(traitSet),
		d: make(traitSet),
		a: make(traitSet),
		v: make(traitSet),
		r: make(traitSet),
	}
	for _, opt := range opts {
		opt(t)
	}
	return &t
}

type traitSet map[string]bool

type traits struct {
	i, t, s, n, d, a, v, r traitSet
}

func (t *traits) Ignore(url string) bool {
	return t.i[url]
}

func (t *traits) IsSymmetric(url string) bool {
	return t.s[url]
}

func (t *traits) IsInverted(url string) bool {
	return t.v[url]
}

func (t *traits) IsTransitive(url string) bool {
	return t.t[url]
}

func (t *traits) IsName(url string) bool {
	return t.n[url]
}

func (t *traits) IsDistinct(url string) bool {
	return t.d[url]
}

func (t *traits) IsAmbiguous(url string) bool {
	return t.a[url]
}

func (t *traits) IsRule(url string) bool {
	return t.r[url]
}
