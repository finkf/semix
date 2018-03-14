// Package resource defines the configuration
// for a knowledge base resource.
// It uses a simple toml file format
package resource

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"bitbucket.org/fflo/semix/pkg/rdfxml"
	"bitbucket.org/fflo/semix/pkg/say"
	"bitbucket.org/fflo/semix/pkg/semix"
	"bitbucket.org/fflo/semix/pkg/traits"
	"bitbucket.org/fflo/semix/pkg/turtle"
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Comparision ignores case
const (
	// RDFXML sets the input format to RDF XML
	RDFXML = "rdfxml"
	// Turtle sets the input format to Turtle
	Turtle = "turtle"
	// Discard sets the ambig handler to discard all ambiguities.
	Discard = "discard"
	// Merge sets the ambig handler to merge ambiguities to a distinct concept.
	Merge = "merge"
	// Split sets the ambig handler to split ambiguities to an ambigiuous concept.
	Split = "split"
	// Fail sets the ambig handle to fail on any internal ambiguity.
	Fail = "fail"
)

type file struct {
	Path, Type, Cache, Ambigs string
	handle                    semix.HandleAmbigsFunc
}

type predicates struct {
	Ignore     []string
	Transitive []string
	Symmetric  []string
	Name       []string
	Distinct   []string
	Ambiguous  []string
	Inverted   []string
	Rule       []string
}

// Config represents the configuration for a knowledge base.
type Config struct {
	File       file
	Predicates predicates
}

// Parse is a convinence fuction that parses a knowledge base
// using a toml configuration file.
func Parse(file string, useCache bool) (*semix.Resource, error) {
	c, err := Read(file)
	if err != nil {
		return nil, err
	}
	return c.Parse(useCache)
}

// Read reads a configuration from a file.
// $VAR and ${VAR} in file.path and file.cache
// are automatically expanded using the environment.
func Read(file string) (*Config, error) {
	var c Config
	if _, err := toml.DecodeFile(file, &c); err != nil {
		return nil, err
	}
	c.File.Cache = os.ExpandEnv(c.File.Cache)
	c.File.Path = os.ExpandEnv(c.File.Path)
	handle, err := c.newHandle()
	if err != nil {
		return nil, err
	}
	c.File.handle = handle
	return &c, nil
}

// Parse parses the configuration and returns the graph and the dictionary.
// If useCache is false, the cache is neither read nor written.
func (c *Config) Parse(useCache bool) (*semix.Resource, error) {
	if useCache && c.File.Cache != "" {
		if r, err := c.readCache(); err == nil {
			return r, nil
		}
	}
	is, err := os.Open(c.File.Path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = is.Close() }()
	parser, err := c.newParser(is)
	if err != nil {
		return nil, err
	}
	r, err := semix.Parse(parser, c.Traits())
	if err != nil {
		return nil, err
	}
	if useCache && c.File.Cache != "" {
		if err := c.writeCache(r); err != nil {
			say.Info("error: %s", err)
		}
	}
	say.Debug("loaded %d concepts, %d entries, %d rules",
		r.Graph.ConceptsLen(), len(r.Dictionary), len(r.Rules))
	return r, nil
}

// Traits returns a new Traits interface using the configuration
// of this config file.
func (c *Config) Traits() semix.Traits {
	return traits.New(
		traits.WithIgnorePredicates(c.Predicates.Ignore...),
		traits.WithTransitivePredicates(c.Predicates.Transitive...),
		traits.WithSymmetricPredicates(c.Predicates.Symmetric...),
		traits.WithNamePredicates(c.Predicates.Name...),
		traits.WithAmbiguousPredicates(c.Predicates.Ambiguous...),
		traits.WithDistinctPredicates(c.Predicates.Distinct...),
		traits.WithInvertedPredicates(c.Predicates.Inverted...),
		traits.WithRulePredicates(c.Predicates.Rule...),
		traits.WithHandleAmbigs(c.File.handle),
	)
}

func (c *Config) newHandle() (semix.HandleAmbigsFunc, error) {
	switch strings.ToLower(c.File.Ambigs) {
	case Merge:
		return semix.HandleAmbigsWithMerge, nil
	case Split:
		return semix.HandleAmbigsWithSplit, nil
	case Discard:
		return func(*semix.Graph, ...string) (*semix.Concept, error) {
			return nil, nil
		}, nil
	case Fail:
		return func(g *semix.Graph, urls ...string) (*semix.Concept, error) {
			return nil, fmt.Errorf("internal ambiguity detected for: %s", urls)
		}, nil
	default:
		t, err := strconv.ParseFloat(c.File.Ambigs, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ambig handler: %s", c.File.Ambigs)
		}
		if t < 0 || t > 1 {
			return nil, fmt.Errorf("invalid ambig handler: %s", c.File.Ambigs)
		}
		return automaticHandleAmbigsFunc(t), nil
	}
}

func (c *Config) newParser(r io.Reader) (semix.Parser, error) {
	switch strings.ToLower(c.File.Type) {
	case RDFXML:
		return rdfxml.NewParser(r), nil
	case Turtle:
		return turtle.NewParser(r), nil
	default:
		return nil, fmt.Errorf("invalid parser type: %s", c.File.Type)
	}
}

func (c *Config) readCache() (*semix.Resource, error) {
	say.Debug("readCache(): %s", c.File.Cache)
	file, err := os.Open(c.File.Cache)
	if err != nil {
		say.Info("error: %s", err)
		return nil, err
	}
	defer func() { _ = file.Close() }()
	r := new(semix.Resource)
	if err := gob.NewDecoder(file).Decode(r); err != nil {
		say.Info("error: %s", err)
		return nil, err
	}
	return r, nil
}

func (c *Config) writeCache(r *semix.Resource) error {
	say.Debug("writeCache(): %s", c.File.Cache)
	if err := os.MkdirAll(filepath.Dir(c.File.Cache), os.ModePerm); err != nil {
		say.Info("error: %s", err)
		return errors.Wrapf(err, "cannot create cache directory")
	}
	file, err := os.Create(c.File.Cache)
	if err != nil {
		say.Info("error: %s", err)
		return errors.Wrapf(err, "cannot write cache")
	}
	defer func() { _ = file.Close() }()
	return gob.NewEncoder(file).Encode(r)
}

func automaticHandleAmbigsFunc(t float64) semix.HandleAmbigsFunc {
	return func(g *semix.Graph, urls ...string) (*semix.Concept, error) {
		min := -1
		say.Debug("handling ambiguity: %s", urls)
		for _, url := range urls {
			c, ok := g.FindByURL(url)
			if !ok {
				continue
			}
			if c.EdgesLen() < min || min == -1 {
				min = c.EdgesLen()
			}
		}
		if min == 0 {
			say.Debug("min=%d: splitting", min)
			return semix.HandleAmbigsWithSplit(g, urls...)
		}
		edges := semix.IntersectEdges(g, urls...)
		var n int
		for _, os := range edges {
			n += len(os)
		}
		if n == 0 {
			say.Debug("min=%d,n=%d: splitting", min, n)
			return semix.HandleAmbigsWithSplit(g, urls...)
		}
		o := float64(n) / float64(min)
		if o < t {
			say.Debug("min=%d,n=%d,%f<%f: splitting", min, n, o, t)
			return semix.HandleAmbigsWithSplit(g, urls...)
		}
		say.Debug("min=%d,n=%d,%f>=%f: merging", min, n, o, t)
		return semix.HandleAmbigsWithMerge(g, urls...)
	}
}
