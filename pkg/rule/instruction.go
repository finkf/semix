package rule

import (
	"fmt"
	"math"
	"sort"

	"bitbucket.org/fflo/semix/pkg/memory"
	"bitbucket.org/fflo/semix/pkg/semix"
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
	opMIN
	opMAX
	opC
	opCS
	opSC
	opSCS
	opE
	opES
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

func (i instruction) call(mem *memory.Memory, stack *stack) {
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
	case opMIN:
		a := stack.popArray1()
		stack.push(arrayMIN(a))
	case opMAX:
		a := stack.popArray1()
		stack.push(arrayMAX(a))
	case opSC:
		a := stack.pop1()
		stack.push(float64(countSC(mem, int(a))))
	case opSCS:
		a := stack.pop1()
		stack.push(float64(countSCS(mem, int(a))))
	case opC:
		a := stack.popArray1()
		stack.pushArray(countCS(mem, a))
	case opCS:
		a := stack.popArray1()
		stack.pushArray(countCS(mem, a))
	case opE:
		stack.pushArray(elems(mem))
	case opES:
		stack.pushArray(elemsS(mem))
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
		return fmt.Sprintf("PUSH %.2f", i.arg)
	case opPushID:
		return fmt.Sprintf("PUSH %d", int(i.arg))
	case opPushTRUE:
		return fmt.Sprintf("PUSH %t", true)
	case opPushFALSE:
		return fmt.Sprintf("PUSH %t", false)
	case opEQ:
		return "EQ"
	case opLT:
		return "LT"
	case opGT:
		return "GT"
	case opNOT:
		return "NOT"
	case opNEG:
		return "NEG"
	case opADD:
		return "ADD"
	case opSUB:
		return "SUB"
	case opMUL:
		return "MUL"
	case opDIV:
		return "DIV"
	case opOR:
		return "OR"
	case opAND:
		return "AND"
	case opSetEQ:
		return "SEQ"
	case opSetU:
		return "SU"
	case opSetSUB:
		return "SSUB"
	case opLEN:
		return "LEN"
	case opLOG:
		return "LOG"
	case opEXP:
		return "EXP"
	case opPOW:
		return "POW"
	case opMIN:
		return "MIN"
	case opMAX:
		return "MAX"
	case opC:
		return "C"
	case opCS:
		return "CS"
	case opSC:
		return "SC"
	case opSCS:
		return "SCS"
	case opE:
		return "E"
	case opES:
		return "ES"
	}
	panic("invalid opcode")
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

func arrayMIN(a []float64) float64 {
	if len(a) == 0 {
		return -math.MaxFloat64
	}
	min := a[0]
	for i := 1; i < len(a); i++ {
		min = math.Min(min, a[i])
	}
	return min
}

func arrayMAX(a []float64) float64 {
	if len(a) == 0 {
		return math.MaxFloat64
	}
	max := a[0]
	for i := 1; i < len(a); i++ {
		max = math.Max(max, a[i])
	}
	return max
}

func elemsS(mem *memory.Memory) []float64 {
	ids := make(map[int]bool, mem.N())
	eachS(mem, func(c *semix.Concept) {
		ids[absID(c.ID())] = true
	})
	return toset(ids)
}

func elems(mem *memory.Memory) []float64 {
	ids := make(map[int]bool, mem.N())
	mem.Each(func(c *semix.Concept) {
		ids[absID(c.ID())] = true
	})
	return toset(ids)
}

func countC(mem *memory.Memory, ids []float64) []float64 {
	counts := make(map[int]bool, mem.N())
	for _, id := range ids {
		mem.Each(func(c *semix.Concept) {
			counts[countSC(mem, int(id))] = true
		})
	}
	return toset(counts)
}

func countCS(mem *memory.Memory, ids []float64) []float64 {
	counts := make(map[int]bool, mem.N())
	for _, id := range ids {
		eachS(mem, func(c *semix.Concept) {
			counts[countSC(mem, int(id))] = true
		})
	}
	return toset(counts)
}

func countSC(mem *memory.Memory, id int) int {
	var count int
	mem.Each(func(c *semix.Concept) {
		if absID(c.ID()) == id {
			count++
		}
	})
	return count
}

func countSCS(mem *memory.Memory, id int) int {
	var count int
	eachS(mem, func(c *semix.Concept) {
		if absID(c.ID()) == id {
			count++
		}
	})
	return count
}

func eachS(mem *memory.Memory, f func(*semix.Concept)) {
	mem.Each(func(c *semix.Concept) {
		f(c)
		c.EachEdge(func(e semix.Edge) {
			f(e.O)
		})
	})
}

func toset(ids map[int]bool) []float64 {
	set := make([]float64, 0, len(ids))
	for id := range ids {
		set = append(set, float64(id))
	}
	sort.Float64s(set)
	return set
}

func absID(id int32) int {
	if id < 0 {
		return -int(id)
	}
	return int(id)
}
