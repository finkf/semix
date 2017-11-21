package rule

import (
	"fmt"
	"strings"
)

type function struct {
	name string
	args []ast
}

func (function) typ() astType {
	return astFunction
}

func (f function) String() string {
	var strs []string
	for _, arg := range f.args {
		strs = append(strs, arg.String())
	}
	return fmt.Sprintf("%s(%s)", f.name, strings.Join(strs, ","))
}

func (f function) check() astType {
	switch f.name {
	case "min":
		return f.minMaxCheck()
	case "max":
		return f.minMaxCheck()
	case "len":
		return f.lenCheck()
	case "es":
		return f.elementsCheck()
	case "e":
		return f.elementsCheck()
	case "c":
		return f.countsCheck()
	case "cs":
		return f.countsCheck()
	default:
		astFatalf("invalid function name: %s", f)
	}
	panic("unreacheable")
}

func (f function) compile(func(string) int) Rule {
	astFatalf("cannot compile %s: not implemented", f)
	panic("unreacheable")
}

func (f function) elementsCheck() astType {
	if len(f.args) != 0 {
		astFatalf("invalid arguments: %s", f)
	}
	return astSet
}

func (f function) countsCheck() astType {
	if len(f.args) == 0 {
		astFatalf("invalid arguments: %s", f)
	}
	for _, arg := range f.args {
		t := arg.check()
		if t != astSet && t != astStr {
			astFatalf("invalid arguments: %s", f)
		}
	}
	return astNumArray
}

func (f function) lenCheck() astType {
	if len(f.args) != 1 {
		astFatalf("invalid arguments: %s", f)
	}
	t := f.args[0].check()
	if t != astSet && t != astStr {
		astFatalf("invalid arguments: %s", f)
	}
	return astNum
}
func (f function) minMaxCheck() astType {
	if len(f.args) == 0 {
		astFatalf("invalid arguments: %s", f)
	}
	t := f.args[0].check()
	if t == astBoolean {
		for i := 1; i < len(f.args); i++ {
			if f.args[i].check() != astBoolean {
				astFatalf("invalid arguments: %s", f)
			}
		}
		return astBoolean
	}
	if t != astNum && t != astNumArray {
		astFatalf("invalid arguments: %s", f)
	}
	for i := 1; i < len(f.args); i++ {
		tt := f.args[i].check()
		if tt != astNum && tt != astNumArray {
			astFatalf("invalid arguments: %s", f)
		}
	}
	return astNum
}
