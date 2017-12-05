package resource

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"bitbucket.org/fflo/semix/pkg/rdfxml"
	"bitbucket.org/fflo/semix/pkg/semix"
	"bitbucket.org/fflo/semix/pkg/traits"
	"bitbucket.org/fflo/semix/pkg/turtle"
	"github.com/BurntSushi/toml"
)

// The parser format identifiers.
const (
	RDFXML = "rdfxml"
	Turtle = "turtle"
)

type file struct {
	Path, Type, Cache string
	Merge             bool
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
	return &c, nil
}

// Parse parses the configuration and returns the graph and the dictionary.
func (c Config) Parse(useCache bool) (*semix.Resource, error) {
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
	if c.File.Cache != "" {
		err := c.writeCache(r)
		if err != nil {
			log.Printf("error: %s", err)
		}
	}
	return r, nil
}

// Traits returns a new Traits interface using the configuration
// of this config file.
func (c Config) Traits() traits.Interface {
	return traits.New(
		traits.WithIgnorePredicates(c.Predicates.Ignore...),
		traits.WithTransitivePredicates(c.Predicates.Transitive...),
		traits.WithSymmetricPredicates(c.Predicates.Symmetric...),
		traits.WithNamePredicates(c.Predicates.Name...),
		traits.WithAmbiguousPredicates(c.Predicates.Ambiguous...),
		traits.WithDistinctPredicates(c.Predicates.Distinct...),
		traits.WithInvertedPredicates(c.Predicates.Inverted...),
		traits.WithRulePredicates(c.Predicates.Rule...),
		traits.WithSplitAmbiguousURLs(!c.File.Merge),
	)
}

func (c Config) newParser(r io.Reader) (semix.Parser, error) {
	switch strings.ToLower(c.File.Type) {
	case RDFXML:
		return rdfxml.NewParser(r), nil
	case Turtle:
		return turtle.NewParser(r), nil
	default:
		return nil, fmt.Errorf("invalid parser type: %s", c.File.Type)
	}
}

func (c Config) readCache() (*semix.Resource, error) {
	log.Printf("readCache(): %s", c.File.Cache)
	file, err := os.Open(c.File.Cache)
	if err != nil {
		log.Printf("error: %s", err)
		return nil, err
	}
	defer file.Close()
	r := new(semix.Resource)
	if err := gob.NewDecoder(file).Decode(r); err != nil {
		log.Printf("error: %s", err)
		return nil, err
	}
	return r, nil
}

func (c Config) writeCache(r *semix.Resource) error {
	log.Printf("writeCache(): %s", c.File.Cache)
	file, err := os.Create(c.File.Cache)
	if err != nil {
		log.Printf("error: %s", err)
		return err
	}
	defer file.Close()
	return gob.NewEncoder(file).Encode(r)
}