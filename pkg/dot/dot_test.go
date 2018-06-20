package dot_test

import (
	"flag"
	"image/png"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"bitbucket.org/fflo/semix/pkg/dot"
)

var (
	update    = flag.Bool("update", false, "update gold file")
	noTestDot = flag.Bool("no-test-dot", false, "disable tests using the dot executable")
)

const (
	goldDotFile = "testdata/gold.dot"
	goldPNGFile = "testdata/gold.png"
	goldSVGFile = "testdata/gold.svg"
	dotExe      = "/usr/bin/dot"
)

func newDot() *dot.Dot {
	d := dot.New("test", dot.Rankdir, dot.LR)
	d.AddNode("1", dot.Label, "Node 1")
	d.AddNode("2", dot.Label, "Node 2")
	d.AddEdge("1", "2", dot.Style, dot.Dotted)
	d.AddEdge("2", "3")
	return d
}

func TestDot(t *testing.T) {
	d := newDot()
	if *update {
		if err := ioutil.WriteFile(goldDotFile, []byte(d.String()), os.ModePerm); err != nil {
			t.Fatalf("cannot update gold file %s: %v", goldDotFile, err)
		}
	}

	gold, err := ioutil.ReadFile(goldDotFile)
	if err != nil {
		t.Fatalf("cannot read gold file %s: %v", goldDotFile, err)
	}
	// split into lines and sort them
	got := d.String()
	goldArr := strings.Split(string(gold), "\n")
	gotArr := strings.Split(got, "\n")
	sort.Strings(goldArr)
	sort.Strings(gotArr)
	if !reflect.DeepEqual(goldArr, gotArr) {
		t.Fatalf("expected %v; got %v", goldArr, gotArr)
	}
}

func TestPNG(t *testing.T) {
	if noTestDot {
		return
	}
	d := newDot()
	if *update {
		img, err := d.PNG(dotExe)
		if err != nil {
			t.Fatalf("cannot generate png: %v", err)
		}
		out, err := os.Create(goldPNGFile)
		if err != nil {
			t.Fatalf("cannot open gold file %s: %v", goldPNGFile, err)
		}
		defer func() { _ = out.Close() }()
		if err := png.Encode(out, img); err != nil {
			t.Fatalf("cannot write gold file %s: %v", goldPNGFile, err)
		}
	}
	in, err := os.Open(goldPNGFile)
	if err != nil {
		t.Fatalf("cannot open gold file %s: %v", goldPNGFile, err)
	}
	defer func() { _ = in.Close() }()
	gold, err := png.Decode(in)
	if err != nil {
		t.Fatalf("cannot read gold file %s: %v", goldPNGFile, err)
	}
	got, err := d.PNG(dotExe)
	if err != nil {
		t.Fatalf("cannot generate png: %v", err)
	}
	if !reflect.DeepEqual(gold, got) {
		t.Fatalf("gold and test images are not the same")
	}
}

func TestSVG(t *testing.T) {
	if *noTestDot {
		return
	}
	d := newDot()
	if *update {
		svg, err := d.SVG(dotExe)
		if err != nil {
			t.Fatalf("cannot generate svg: %v", err)
		}
		out, err := os.Create(goldSVGFile)
		if err != nil {
			t.Fatalf("cannot open gold file %s: %v", goldSVGFile, err)
		}
		defer func() { _ = out.Close() }()
		if err := ioutil.WriteFile(goldSVGFile, []byte(svg), os.ModePerm); err != nil {
			t.Fatalf("cannot write gold file %s: %v", goldSVGFile, err)
		}
	}
	gold, err := ioutil.ReadFile(goldSVGFile)
	if err != nil {
		t.Fatalf("cannot read gold file %s: %v", goldSVGFile, err)
	}
	got, err := d.SVG(dotExe)
	if err != nil {
		t.Fatalf("cannot generate svg: %v", err)
	}
	if got != string(gold) {
		t.Fatalf("expected %s; got %s", gold, got)
	}
}

func TestOddArgs(t *testing.T) {
	tests := []func(){
		func() { d := dot.New("invalid", "a", "b", "c"); d.AddNode("invalid") },
		func() { d := dot.New("test"); d.AddNode("invalid", "a", "b", "c") },
		func() { d := dot.New("test"); d.AddEdge("invalid1", "invalid2", "a") },
	}
	for _, tc := range tests {
		t.Run("panic", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()
			tc()
		})
	}
}
