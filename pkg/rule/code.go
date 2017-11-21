package rule

import "fmt"

type stack []float64

func (s *stack) pop1() float64 {
	n := len(*s)
	res := (*s)[n-1]
	*s = (*s)[:n-1]
	return res
}

func (s *stack) pop2() (float64, float64) {
	a := s.pop1()
	b := s.pop1()
	// switch arguments
	return b, a
}

func (s *stack) popBool1() bool {
	return !(s.pop1() == 0)
}

func (s *stack) popBool2() (bool, bool) {
	a := s.popBool1()
	b := s.popBool1()
	// switch arguments
	return b, a
}

func (s *stack) push(x float64) {
	*s = append(*s, x)
}

func (s *stack) pushBool(b bool) {
	if b {
		*s = append(*s, 1)
	} else {
		*s = append(*s, 0)
	}
}

type opcode int

const (
	opPushNUM opcode = iota
	opPushID
	opPushTRUE
	opPushFALSE
	opEQ
	opLT
	opGT
	opNOT
	opNEG
	opADD
	opSUB
	opDIV
	opMUL
	opOR
	opAND
)

type instruction struct {
	arg    float64
	opcode opcode
}

func booleanInstruction(b bool) instruction {
	if b {
		return instruction{opcode: opPushTRUE}
	}
	return instruction{opcode: opPushFALSE}
}

func (i instruction) call(stack *stack) {
	switch i.opcode {
	case opPushNUM:
		stack.push(i.arg)
	case opPushID:
		panic("opPushID: not implemented")
	case opPushTRUE:
		stack.pushBool(true)
	case opPushFALSE:
		stack.pushBool(false)
	case opEQ:
		a, b := stack.pop2()
		stack.pushBool(a == b)
	case opLT:
		a, b := stack.pop2()
		stack.pushBool(a < b)
	case opGT:
		a, b := stack.pop2()
		stack.pushBool(a > b)
	case opNOT:
		stack.pushBool(!stack.popBool1())
	case opNEG:
		stack.push(-stack.pop1())
	case opADD:
		a, b := stack.pop2()
		stack.push(a + b)
	case opSUB:
		a, b := stack.pop2()
		stack.push(a - b)
	case opMUL:
		a, b := stack.pop2()
		stack.push(a * b)
	case opDIV:
		a, b := stack.pop2()
		stack.push(a / b)
	case opOR:
		a, b := stack.popBool2()
		stack.pushBool(a || b)
	case opAND:
		a, b := stack.popBool2()
		stack.pushBool(a && b)
	default:
		panic("invalid opcode")
	}
}

func (i instruction) String() string {
	switch i.opcode {
	case opPushNUM:
		return fmt.Sprintf("opPUSH(%.2f)", i.arg)
	case opPushID:
		return fmt.Sprintf("opPUSH(%d)", int(i.arg))
	case opPushTRUE:
		return fmt.Sprintf("opPUSH(%t)", true)
	case opPushFALSE:
		return fmt.Sprintf("opPUSH(%t)", false)
	case opEQ:
		return "opEQ"
	case opLT:
		return "opLT"
	case opGT:
		return "opGT"
	case opNOT:
		return "opNOT"
	case opNEG:
		return "opNEG"
	case opADD:
		return "opADD"
	case opSUB:
		return "opSUB"
	case opMUL:
		return "opMUL"
	case opDIV:
		return "opDIV"
	case opOR:
		return "opOR"
	case opAND:
		return "opAND"
	default:
		panic("invalid opcode")
	}
}
