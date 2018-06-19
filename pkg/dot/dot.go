package dot

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os/exec"

	"github.com/pkg/errors"
)

// Various configuration variables for dot graphs.
const (
	Rankdir = "rankdir"
	LR      = "LR"
	TB      = "TB"
	RL      = "RL"
	BT      = "BT"
	Label   = "label"
	Style   = "style"
	Dotted  = "dotted"
	Dashed  = "dashed"
)

// Dot is used to build a dot graph.
type Dot struct {
	name  string
	args  []string
	edges map[edge][]string
	nodes map[string][]string
}

type edge struct {
	id1, id2 string
}

// New create a new dot graph with the given name.
// The function panics if the given arguments are not even numbered.
func New(name string, args ...string) *Dot {
	assertIsEven(args...)
	return &Dot{
		name:  name,
		args:  args,
		nodes: make(map[string][]string),
		edges: make(map[edge][]string),
	}
}

// AddNode adds a node with the given id and the given arguments.
// The function panics if the given arguments are not even numbered.
// If the given node already exists, it is overwritten.
func (d *Dot) AddNode(id string, args ...string) *Dot {
	assertIsEven(args...)
	d.nodes[id] = args
	return d
}

// AddEdge adds an edge to the dot graph.
// The function panics if the given arguments are not even numbered.
func (d *Dot) AddEdge(id1, id2 string, args ...string) *Dot {
	assertIsEven(args...)
	d.edges[edge{id1, id2}] = args
	return d
}

// Returns the string representation of the dot graph.
// The string representation of the dot graph is
// the dot-code to build this graph.
func (d *Dot) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("digraph ")
	buffer.WriteString(d.name)
	buffer.WriteString(" {\n")
	// write args
	for i := 0; i < len(d.args); i += 2 {
		buffer.WriteString(keyval(d.args[i:]))
		buffer.WriteString("\n")
	}
	// write nodes
	for id, args := range d.nodes {
		buffer.WriteString(fmt.Sprintf("%q [", id))
		for i := 0; i < len(args); i += 2 {
			buffer.WriteString(keyval(args[i:]))
		}
		buffer.WriteString("]\n")
	}
	// write edges
	for edge, args := range d.edges {
		buffer.WriteString(fmt.Sprintf("%q -> %q [", edge.id1, edge.id2))
		for i := 0; i < len(args); i += 2 {
			buffer.WriteString(keyval(args[i:]))
		}
		buffer.WriteString("]\n")
	}
	buffer.WriteString("}\n")
	return buffer.String()
}

// PNG converts the dot-graph to a png using the dot-code of
// the graph. It uses the given path to the dot executable.
func (d *Dot) PNG(exe string) (image.Image, error) {
	dot := bytes.NewBufferString(d.String()) // redundant
	out := &bytes.Buffer{}
	cmd := exec.Command(exe, "-Tpng")
	cmd.Stdin = dot
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "cannot execute: %s", exe)
	}
	img, err := png.Decode(out)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read image")
	}
	return img, nil
}

func keyval(kv []string) string {
	return fmt.Sprintf("%s=%q;", kv[0], kv[1])
}

func assertIsEven(args ...string) {
	if len(args)%2 != 0 {
		panic(fmt.Sprintf("odd number of arguments given: %d", len(args)))
	}
}
