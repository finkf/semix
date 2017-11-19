package rule

type precedence int

const (
	lowest  precedence = iota + 1
	equals             // =
	compare            // <, >
	line               // +, -
	dot                // *, /
	neg                // !,-
	call               // func()
)

type operator rune

const (
	eq    operator = '='
	gt    operator = '>'
	lt    operator = '<'
	div   operator = '/'
	mul   operator = '*'
	plus  operator = '+'
	minus operator = '-'
	bang  operator = '!'
)

func (o operator) precedence() precedence {
	switch o {
	case bang:
		return neg
	case eq:
		return equals
	case gt:
		return compare
	case lt:
		return compare
	case div:
		return dot
	case mul:
		return dot
	case plus:
		return line
	case minus:
		return line
	}
	return lowest
}
