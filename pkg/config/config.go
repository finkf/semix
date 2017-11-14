package config

import (
	"fmt"
	"io"
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

type parser struct {
	File, Type string
}

type urls struct {
	Ignore     []string
	Transitive []string
	Symmetric  []string
	Name       []string
	Distinct   []string
	Ambiguous  []string
	Inverted   []string
}

// Config represents the configuration for a knowledge base.
type Config struct {
	Parser parser
	URLs   urls
}

// Parse is a convinence fuction that parses a knowledge base
// using a toml configuration file.
func Parse(file string) (*semix.Graph, semix.Dictionary, error) {
	c, err := Read(file)
	if err != nil {
		return nil, nil, err
	}
	return c.Parse()
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
func (c Config) Parse() (*semix.Graph, semix.Dictionary, error) {
	is, err := os.Open(c.Parser.File)
	if err != nil {
		return nil, nil, err
	}
	defer is.Close()
	parser, err := c.newParser(is)
	if err != nil {
		return nil, nil, err
	}
	return semix.Parse(parser, c.traits())
}

func (c Config) traits() semix.Traits {
	return traits.New(
		traits.WithIgnoreURLs(c.URLs.Ignore...),
		traits.WithTransitiveURLs(c.URLs.Transitive...),
		traits.WithSymmetricURLs(c.URLs.Symmetric...),
		traits.WithNameURLs(c.URLs.Name...),
		traits.WithAmbiguousURLs(c.URLs.Ambiguous...),
		traits.WithDistinctURLs(c.URLs.Distinct...),
		traits.WithInvertedURLs(c.URLs.Inverted...),
	)
}

func (c Config) newParser(r io.Reader) (semix.Parser, error) {
	switch strings.ToLower(c.Parser.Type) {
	case RDFXML:
		return rdfxml.NewParser(r), nil
	case Turtle:
		return turtle.NewParser(r), nil
	default:
		return nil, fmt.Errorf("invalid file type: %s", c.Parser.Type)
	}
}
