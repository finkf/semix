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
	optPushNum = iota
	optPushID
	optPushTrue
	optPushFalse
	optEQ
	optLT
	optGT
	optNot
	optNeg
	optAdd
	optSub
	optDiv
	optMul
)

type optcode struct {
	arg  float64
	code int
}

func (o optcode) call(stack *stack) {
	switch o.code {
	case optPushNum:
		stack.push(o.arg)
	case optPushID:
		panic("optPushID: not implemented")
	case optPushTrue:
		stack.pushBool(true)
	case optPushFalse:
		stack.pushBool(false)
	case optEQ:
		a, b := stack.pop2()
		stack.pushBool(a == b)
	case optLT:
		a, b := stack.pop2()
		stack.pushBool(a < b)
	case optGT:
		a, b := stack.pop2()
		stack.pushBool(a > b)
	case optNot:
		stack.pushBool(!stack.popBool())
	case optNeg:
		stack.push(-stack.pop1())
	case optAdd:
		a, b := stack.pop2()
		stack.push(a + b)
	case optSub:
		a, b := stack.pop2()
		stack.push(a - b)
	case optMul:
		a, b := stack.pop2()
		stack.push(a * b)
	case optDiv:
		a, b := stack.pop2()
		stack.push(a / b)
	default:
		panic("invalid opt code")
	}
}

func (o optcode) String() string {
	switch o.code {
	case optPushNum:
		return fmt.Sprintf("push(%.2f)", o.arg)
	case optPushID:
		return fmt.Sprintf("push(%d)", int(o.arg))
	case optPushTrue:
		return fmt.Sprintf("push(%t)", true)
	case optPushFalse:
		return fmt.Sprintf("push(%t)", false)
	case optEQ:
		return "optEQ"
	case optLT:
		return "optLT"
	case optGT:
		return "optGT"
	case optNot:
		return "optNot"
	case optNeg:
		return "optNeg"
	case optAdd:
		return "optAdd"
	case optSub:
		return "optSub"
	case optMul:
		return "optMul"
	case optDiv:
		return "optDiv"
	default:
		panic("invalid opt code")
	}
}
