package rule

import "fmt"

type stack []float64

func (s *stack) pop1() float64 {
	n := len(*s)
	res := (*s)[n-1]
	*s = (*s)[:n-1]
	return res
}

func (s *stack) popBool() bool {
	return !(s.pop1() == 0)
}

func (s *stack) pop2() (float64, float64) {
	a := s.pop1()
	b := s.pop1()
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

const (
	opPusNum = iota
	opPushID
	opPushTrue
	opPushFalse
	opEQ
	opLT
	opGT
	opNot
	opNeg
	opAdd
	opSub
	opDiv
	opMul
)

type instruction struct {
	arg    float64
	opcode int
}

func (i instruction) call(stack *stack) {
	switch i.opcode {
	case opPusNum:
		stack.push(i.arg)
	case opPushID:
		panic("opPushID: not implemented")
	case opPushTrue:
		stack.pushBool(true)
	case opPushFalse:
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
	case opNot:
		stack.pushBool(!stack.popBool())
	case opNeg:
		stack.push(-stack.pop1())
	case opAdd:
		a, b := stack.pop2()
		stack.push(a + b)
	case opSub:
		a, b := stack.pop2()
		stack.push(a - b)
	case opMul:
		a, b := stack.pop2()
		stack.push(a * b)
	case opDiv:
		a, b := stack.pop2()
		stack.push(a / b)
	default:
		panic("invalid opcode")
	}
}

func (i instruction) String() string {
	switch i.opcode {
	case opPusNum:
		return fmt.Sprintf("opPush(%.2f)", i.arg)
	case opPushID:
		return fmt.Sprintf("opPush(%d)", int(i.arg))
	case opPushTrue:
		return fmt.Sprintf("opPush(%t)", true)
	case opPushFalse:
		return fmt.Sprintf("opPush(%t)", false)
	case opEQ:
		return "opEQ"
	case opLT:
		return "opLT"
	case opGT:
		return "opGT"
	case opNot:
		return "opNot"
	case opNeg:
		return "opNeg"
	case opAdd:
		return "opAdd"
	case opSub:
		return "opSub"
	case opMul:
		return "opMul"
	case opDiv:
		return "opDiv"
	default:
		panic("invalid opcode")
	}
}
