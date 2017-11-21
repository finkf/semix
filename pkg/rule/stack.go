package rule

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

func (s *stack) popArray1() []float64 {
	n := int(s.pop1())
	n = len(*s) - n
	a := (*s)[n:]
	*s = (*s)[:n]
	return a
}

func (s *stack) popArray2() ([]float64, []float64) {
	a := s.popArray1()
	b := s.popArray1()
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

func (s *stack) pushArray(a []float64) {
	*s = append(*s, a...)
	*s = append(*s, float64(len(a)))
}
