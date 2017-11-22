package rule

import (
	"fmt"
)

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
	}
	astFatalf("invalid expression: %s", i)
	panic("unreacheable")
}

func (i infix) compile(f func(string) int) Rule {
	switch i.left.check() {
	case astBoolean:
		switch i.op {
		case '=':
			return i.combine(f, instruction{opcode: opEQ})
		case '+':
			return i.combine(f, instruction{opcode: opOR})
		case '*':
			return i.combine(f, instruction{opcode: opAND})
		}
	case astNum:
		switch i.op {
		case '=':
			return i.combine(f, instruction{opcode: opEQ})
		case '<':
			return i.combine(f, instruction{opcode: opLT})
		case '>':
			return i.combine(f, instruction{opcode: opGT})
		case '+':
			return i.combine(f, instruction{opcode: opADD})
		case '-':
			return i.combine(f, instruction{opcode: opSUB})
		case '*':
			return i.combine(f, instruction{opcode: opMUL})
		case '/':
			return i.combine(f, instruction{opcode: opDIV})
		}
	case astSet:
		switch i.op {
		case '=':
			return i.combine(f, instruction{opcode: opSetEQ})
		case '+':
			return i.combine(f, instruction{opcode: opSetU})
		case '*':
			return i.combine(f, instruction{opcode: opSetI})
		case '-':
			return i.combine(f, instruction{opcode: opSetSUB})
		}
	case astStr:
		switch i.op {
		case '=':
			return Rule{booleanInstruction(i.left.(str) == i.right.(str))}
		case '<':
			return Rule{booleanInstruction(i.left.(str) < i.right.(str))}
		case '>':
			return Rule{booleanInstruction(i.left.(str) > i.right.(str))}
		}
	}
	astFatalf("invalid type or operator: %s", i)
	panic("unreacheable")
}

func (i infix) combine(f func(string) int, instr instruction) Rule {
	rule := i.left.compile(f)
	rule = append(rule, i.right.compile(f)...)
	return append(rule, instr)

}

func (i infix) String() string {
	return fmt.Sprintf("(%s%c%s)", i.left, i.op, i.right)
}
