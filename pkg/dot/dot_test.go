package dot_test

import (
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"bitbucket.org/fflo/semix/pkg/dot"
)

var update = flag.Bool("update", false, "update gold file")

const goldFile = "testdata/gold.dot"

func TestDot(t *testing.T) {
	d := dot.New("test", dot.Rankdir, dot.LR)
	d.AddNode("1", dot.Label, "Node 1")
	d.AddNode("2", dot.Label, "Node 2")
	d.AddEdge("1", "2", dot.Style, dot.Dotted)
	d.AddEdge("2", "3")

	if *update {
		if err := ioutil.WriteFile(goldFile, []byte(d.String()), os.ModePerm); err != nil {
			t.Fatalf("cannot update gold file %s: %v", goldFile, err)
		}
	}

	gold, err := ioutil.ReadFile(goldFile)
	if err != nil {
		t.Fatalf("cannot read gold file %s: %v", goldFile, err)
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
