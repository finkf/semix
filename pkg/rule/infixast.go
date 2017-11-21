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

func (i infix) compile(func(string) int) Rule {
	switch i.left.check() {
	case astStr:
		switch i.op {
		case '=':
			return i.compileStrEQ()
		default:
			panic("invalid operator")
		}
	// case astNum:
	// case astBoolean:
	// case astSet:
	default:
		panic("invalid type")
	}
	// astFatalf("cannot compile %s: not implemented", i)
	// panic("unreacheable")
}

func (i infix) compileStrEQ() Rule {
	b := i.left.(str) == i.right.(str)
	if b {
		return Rule{instruction{opcode: opPushTrue}}
	}
	return Rule{instruction{opcode: opPushFalse}}
}

func (i infix) String() string {
	return fmt.Sprintf("(%s%c%s)", i.left, i.op, i.right)
}
