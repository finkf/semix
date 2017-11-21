package rule

import (
	"fmt"
	"math"
)

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
	opSetEQ
	opSetU
	opSetI
	opSetSUB
	opLEN
	opLOG
	opEXP
	opPOW
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
		stack.push(i.arg)
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
	case opSetEQ:
		a, b := stack.popArray2()
		stack.pushBool(arrayEQ(a, b))
	case opSetU:
		a, b := stack.popArray2()
		stack.pushArray(arrayU(a, b))
	case opSetI:
		a, b := stack.popArray2()
		stack.pushArray(arrayI(a, b))
	case opSetSUB:
		a, b := stack.popArray2()
		stack.pushArray(arraySUB(a, b))
	case opLEN:
		a := stack.popArray1()
		stack.push(float64(len(a)))
	case opLOG:
		a := stack.pop1()
		stack.push(math.Log(a))
	case opEXP:
		a := stack.pop1()
		stack.push(math.Exp(a))
	case opPOW:
		a, b := stack.pop2()
		stack.push(math.Pow(a, b))
	default:
		panic("invalid opcode")
	}
}

func arrayEQ(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
	case opSetEQ:
		return "opSetEQ"
	case opSetU:
		return "opSetU"
	case opSetSUB:
		return "opSetSUB"
	case opLEN:
		return "opLEN"
	case opLOG:
		return "opLOG"
	case opEXP:
		return "opEXP"
	case opPOW:
		return "opPOW"
	default:
		panic("invalid opcode")
	}
}

func arrayU(a, b []float64) []float64 {
	res := make([]float64, 0, len(a)+len(b))
	var i, j int
	for i, j = 0, 0; i < len(a) && j < len(b); {
		if a[i] < b[j] {
			res = append(res, a[i])
			i++
		} else if b[j] < a[i] {
			res = append(res, b[j])
			j++
		} else {
			res = append(res, a[i])
			i++
			j++
		}
	}
	for ; i < len(a); i++ {
		res = append(res, a[i])
	}
	for ; j < len(b); j++ {
		res = append(res, b[j])
	}
	return res
}

func arrayI(a, b []float64) []float64 {
	res := make([]float64, 0, (len(a)+len(b))/2)
	var i, j int
	for i, j = 0, 0; i < len(a) && j < len(b); {
		if a[i] < b[j] {
			i++
		} else if b[j] < a[i] {
			j++
		} else {
			res = append(res, a[i])
			i++
			j++
		}
	}
	return res
}

func arraySUB(a, b []float64) []float64 {
	res := make([]float64, 0, len(a))
	var i, j int
	for i, j = 0, 0; i < len(a) && j < len(b); {
		if a[i] < b[j] {
			res = append(res, a[i])
			i++
		} else if b[j] < a[i] {
			j++
		} else {
			i++
			j++
		}
	}
	for ; i < len(a); i++ {
		res = append(res, a[i])
	}
	return res
}
