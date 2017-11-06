package semix

// ExpandBraces expands braces in a given string using a bash-like syntax.
func ExpandBraces(str string) ([]string, error) {
	e := expander{str: str}
	return e.parse(0), nil
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

func (e *expander) parse(bcount int) []string {
	s := []string{""}
	for c, esc := e.next(); c != 0; c, esc = e.next() {
		l := len(s) - 1
		switch {
		case esc:
			if bcount > 0 {
				s[l] += string(c)
			} else {
				for i := range s {
					s[i] += string(c)
				}
			}
		case c == ',':
			if bcount > 0 {
				s = append(s, "")
			} else {
				s[l] += string(c)
			}
		case c == '{':
			ss := e.parse(bcount + 1)
			s = combine(s, ss)
		case c == '}':
			return s
		default:
			if bcount > 0 {
				s[l] += string(c)
			} else {
				for i := range s {
					s[i] += string(c)
				}
			}
		}
	}
	return s
}

func combine(a, b []string) []string {
	var res []string
	for _, astr := range a {
		for _, bstr := range b {
			res = append(res, astr+bstr)
		}
	}
	return res
}
