package rule

import "fmt"

type infix struct {
	op          operator
	left, right ast
}

func (infix) typ() astType {
	return astInfix
}

func (i infix) check() astType {
	left := i.left.check()
	right := i.right.check()
	if left != right {
		astFatalf("invalid expression: %s", i)
	}
	switch i.op {
	case '=':
		checkTypIn(i, left, astBoolean, astNum, astSet, astStr)
		return astBoolean
	case '>':
		checkTypIn(i, left, astNum, astStr)
		return astBoolean
	case '<':
		checkTypIn(i, left, astNum, astStr)
		return astBoolean
	case '+':
		checkTypIn(i, left, astBoolean, astNum, astSet, astStr)
		return left
	case '-':
		checkTypIn(i, left, astNum, astSet)
		return left
	case '/':
		checkTypIn(i, left, astNum)
		return left
	case '*':
		checkTypIn(i, left, astBoolean, astNum, astSet)
		return left
	default:
		astFatalf("invalid expression: %s", i)
	}
	panic("unreacheable")
}

func (i infix) compile(f func(string) int) Rule {
	switch i.left.check() {
	case astBoolean:
		switch i.op {
		case '=':
			return i.compileBooleanEQ(f)
		case '+':
			return i.compileBooleanOR(f)
		case '*':
			return i.compileBooleanAND(f)
		default:
			panic("invalid operator")
		}
	case astStr:
		switch i.op {
		case '=':
			return i.compileStrEQ()
		case '<':
			return i.compileStrLT()
		case '>':
			return i.compileStrGT()
		default:
			panic("invalid operator")
		}
	// case astNum:
	// case astSet:
	default:
		panic("invalid type")
	}
	// astFatalf("cannot compile %s: not implemented", i)
	// panic("unreacheable")
}

func (i infix) compileBooleanEQ(f func(string) int) Rule {
	return i.combine(f, instruction{opcode: opEQ})
}

func (i infix) compileBooleanOR(f func(string) int) Rule {
	return i.combine(f, instruction{opcode: opOR})
}

func (i infix) compileBooleanAND(f func(string) int) Rule {
	return i.combine(f, instruction{opcode: opAND})
}

func (i infix) compileStrEQ() Rule {
	return Rule{booleanInstruction(i.left.(str) == i.right.(str))}
}

func (i infix) compileStrLT() Rule {
	return Rule{booleanInstruction(i.left.(str) < i.right.(str))}
}

func (i infix) compileStrGT() Rule {
	return Rule{booleanInstruction(i.left.(str) > i.right.(str))}
}

func (i infix) combine(f func(string) int, instr instruction) Rule {
	rule := i.left.compile(f)
	rule = append(rule, i.right.compile(f)...)
	return append(rule, instr)

}

func (i infix) String() string {
	return fmt.Sprintf("(%s%c%s)", i.left, i.op, i.right)
}
