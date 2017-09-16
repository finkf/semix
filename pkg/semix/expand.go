package semix

// ExpandBraces expands braces in a given string using a bash-like syntax.
func ExpandBraces(str string) ([]string, error) {
	e := expander{str: str}
	return e.parse(), nil
}

type expander struct {
	str string
	pos int
}

func (e *expander) next() (byte, bool) {
	if e.pos >= len(e.str) {
		return 0, false
	}
	if e.str[e.pos] == '\\' {
		e.pos++
		if e.pos >= len(e.str) {
			return 0, false
		}
		pos := e.pos
		e.pos++
		if e.str[pos] == '\\' {
			return '\\', false
		}
		return e.str[pos], true
	}
	pos := e.pos
	e.pos++
	return e.str[pos], false
}

func (e *expander) parse() []string {
	s := []string{""}
	for c, esc := e.next(); c != 0; c, esc = e.next() {
		l := len(s) - 1
		switch {
		case esc:
			s[l] += string(c)
		case c == ',':
			s = append(s, "")
		case c == '{':
			ss := e.parse()
			s = combine(s, ss)
		case c == '}':
			return s
		default:
			s[l] += string(c)
		}
	}
	return s
}

func combine(a, b []string) []string {
	if len(a[len(a)-1]) == 0 {
		a = a[:len(a)-1]
		return append(a, b...)
	}
	var res []string
	for _, astr := range a {
		for _, bstr := range b {
			res = append(res, astr+bstr)
		}
	}
	return res
}
