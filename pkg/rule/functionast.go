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

func (f function) compile(l func(string) int) Rule {
	switch f.name {
	case "len":
		switch f.args[0].check() {
		case astStr:
			return Rule{instruction{opcode: opPushID, arg: float64(len(f.args[0].(str)))}}
		case astSet:
			return append(f.args[0].compile(l), instruction{opcode: opLEN})
		default:
			panic("invalid arg type")
		}
	}
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
	if len(f.args) != 1 {
		astFatalf("invalid arguments: %s", f)
	}
	switch f.args[0].check() {
	case astSet:
		return astSet
	case astStr:
		return astNum
	default:
		astFatalf("invalid arguments: %s", f)
	}
	panic("unreacheable")
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
	for _, arg := range f.args {
		t := arg.check()
		if t != astBoolean && t != astNum && t != astSet {
			astFatalf("invalid arguments: %s", f)
		}
	}
	return astNum
}
