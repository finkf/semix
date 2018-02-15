// Package resource defines the configuration
// for a knowledge base resource.
// It uses a simple toml file format
package resource

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"bitbucket.org/fflo/semix/pkg/rdfxml"
	"bitbucket.org/fflo/semix/pkg/semix"
	"bitbucket.org/fflo/semix/pkg/traits"
	"bitbucket.org/fflo/semix/pkg/turtle"
	"github.com/BurntSushi/toml"
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
func Read(file string) (*Config, error) {
	var c Config
	// c := new(Config)
	if _, err := toml.DecodeFile(file, &c); err != nil {
		return nil, err
	}
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
	defer is.Close()
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
			log.Printf("error: %s", err)
		}
	}
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
		return func(*semix.Graph, ...string) *semix.Concept {
			return nil
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
	log.Printf("readCache(): %s", c.File.Cache)
	file, err := os.Open(c.File.Cache)
	if err != nil {
		log.Printf("error: %s", err)
		return nil, err
	}
	defer func() { _ = file.Close() }()
	r := new(semix.Resource)
	if err := gob.NewDecoder(file).Decode(r); err != nil {
		log.Printf("error: %s", err)
		return nil, err
	}
	return r, nil
}

func (c *Config) writeCache(r *semix.Resource) error {
	log.Printf("writeCache(): %s", c.File.Cache)
	file, err := os.Create(c.File.Cache)
	if err != nil {
		log.Printf("error: %s", err)
		return err
	}
	defer func() { _ = file.Close() }()
	return gob.NewEncoder(file).Encode(r)
}

func automaticHandleAmbigsFunc(t float64) semix.HandleAmbigsFunc {
	return func(g *semix.Graph, urls ...string) *semix.Concept {
		min := -1
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
			return semix.HandleAmbigsWithSplit(g, urls...)
		}
		edges := semix.IntersectEdges(g, urls...)
		var n int
		for _, os := range edges {
			n += len(os)
		}
		if n == 0 {
			return semix.HandleAmbigsWithSplit(g, urls...)
		}
		o := float64(n) / float64(min)
		if o < t {
			return semix.HandleAmbigsWithSplit(g, urls...)
		}
		return semix.HandleAmbigsWithMerge(g, urls...)
	}
}
