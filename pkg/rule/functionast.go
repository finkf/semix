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
	case "log":
		return f.numCheck(1)
	case "exp":
		return f.numCheck(1)
	case "pow":
		return f.numCheck(2)
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
			return Rule{
				instruction{opcode: opPushID, arg: float64(len(f.args[0].(str)))},
			}
		case astSet:
			return append(f.combine(l, f.args[0]), instruction{opcode: opLEN})
		}
	case "max":
		return append(f.minMaxCombine(l), instruction{opcode: opMAX})
	case "min":
		return append(f.minMaxCombine(l), instruction{opcode: opMIN})
	case "log":
		return append(f.combine(l, f.args...), instruction{opcode: opLOG})
	case "exp":
		return append(f.combine(l, f.args...), instruction{opcode: opEXP})
	case "pow":
		return append(f.combine(l, f.args...), instruction{opcode: opPOW})
	}
	astFatalf("cannot compile %s: invalid type or instruction")
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
	if len(f.args) == 1 && f.args[0].check() == astSet {
		return astSet
	}
	for _, arg := range f.args {
		if arg.check() != astStr {
			astFatalf("invalid arguments: %s", f)
		}
	}
	return astSet
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
	if len(f.args) == 1 && f.args[0].check() == astSet {
		return astNum
	}
	for _, arg := range f.args {
		t := arg.check()
		if t != astBoolean && t != astNum {
			astFatalf("invalid arguments: %s", f)
		}
	}
	return astNum
}

func (f function) numCheck(n int) astType {
	if len(f.args) != n {
		astFatalf("invalid arguments: %s", f)
	}
	for i := 0; i < n; i++ {
		if f.args[i].check() != astNum {
			astFatalf("invalid arguments: %s", f)
		}
	}
	return astNum
}

func (f function) minMaxCombine(g func(string) int) Rule {
	if len(f.args) == 0 {
		return Rule{instruction{opcode: opPushID, arg: 0}}
	}
	var rule Rule
	var n int
	for _, arg := range f.args {
		if arg.check() == astSet {
			return arg.compile(g)
		}
		n++
		rule = append(rule, arg.compile(g)...)
	}
	rule = append(rule, instruction{opcode: opPushID, arg: float64(n)})
	return rule
}

func (f function) combine(g func(string) int, args ...ast) Rule {
	var rule Rule
	for _, arg := range args {
		rule = append(rule, arg.compile(g)...)
	}
	return rule
}
