package rule

import (
	"fmt"
	"sort"
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
	case "e":
		return Rule{instruction{opcode: opE}}
	case "es":
		return Rule{instruction{opcode: opES}}
	case "c":
		return f.compileCount(l, false)
	case "cs":
		return f.compileCount(l, true)
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

// c/cs is overloaded. It accepts either one string and returns its count,
// or it accepts a set and returns the array of its counts.
func (f function) countsCheck() astType {
	if len(f.args) != 1 {
		astFatalf("invalid arguments: %s", f)
	}
	switch f.args[0].check() {
	case astSet:
		return astSet
	case astStr:
		return astNum
	}
	astFatalf("invalid arguments: %s", f)
	panic("unreacheable")
}

func (f function) compileCount(g func(string) int, star bool) Rule {
	switch f.args[0].check() {
	case astStr:
		return Rule{
			instruction{opcode: opPushID, arg: mustFindID(f.args[0], g)},
			instruction{opcode: countOpcode(true, star)},
		}
	case astSet:
		var rule Rule
		for _, id := range mustFindIDs(f.args[0], g) {
			rule = append(rule, instruction{opcode: opPushID, arg: id})
		}
		rule = append(rule, instruction{opcode: opPushID, arg: float64(len(rule))})
		rule = append(rule, instruction{opcode: countOpcode(false, star)})
		return rule
	}
	astFatalf("invalid arguments: %s", f)
	panic("unreacheable")
}

func countOpcode(str, star bool) opcode {
	if str && star {
		return opSCS
	}
	if str {
		return opSC
	}
	if !str && !star {
		return opC
	}
	return opCS
}

func mustFindIDs(ast ast, f func(string) int) []float64 {
	var ids []float64
	for str := range ast.(set) {
		ids = append(ids, mustFindID(str, f))
	}
	sort.Float64s(ids)
	return ids
}

func mustFindID(ast ast, f func(string) int) float64 {
	id := f(string(ast.(str)))
	if id <= 0 {
		astFatalf("cannot find concept: %s", ast)
	}
	return float64(id)
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
