package semix

import "fmt"

// ExpandBraces expands braces in a given string using a bash-like syntax.
func ExpandBraces(str string) ([]string, error) {
	e := expander{str: str}
	return e.parse(0)
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

func (e *expander) parse(bcount int) ([]string, error) {
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
			ss, err := e.parse(bcount + 1)
			if err != nil {
				return nil, err
			}
			s = combine(s, ss)
		case c == '}':
			if bcount <= 0 {
				return nil, fmt.Errorf("invalid expansion: %s: unbalanced bracets", e.str)
			}
			return s, nil
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
	if bcount > 0 {
		return nil, fmt.Errorf("invalid expansion: %s: unbalanced bracets", e.str)
	}
	return s, nil
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
