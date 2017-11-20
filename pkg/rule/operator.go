package rule

const (
	lowest  = iota + 1
	equals  // =
	compare // <, >
	line    // +, -
	dot     // *, /
	neg     // !,-
	call    // func()
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

func precedence(tok rune) int {
	switch tok {
	case '!':
		return neg
	case '=':
		return equals
	case '>':
		return compare
	case '<':
		return compare
	case '/':
		return dot
	case '*':
		return dot
	case '+':
		return line
	case '-':
		return line
	}
	return lowest
}
