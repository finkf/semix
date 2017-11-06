package semix

import (
	"fmt"
	"log"
)

// ExpandBraces expands braces in a given string using a bash-like syntax.
func ExpandBraces(str string) ([]string, error) {
	e := expander{str: []byte(str)}
	var res []string
	bss, err := e.parse(0)
	if err != nil {
		return nil, err
	}
	for _, bs := range bss {
		res = append(res, string(bs))
	}
	return res, nil
}

type expander struct {
	str []byte
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

func (e *expander) parse(bcount int) ([][]byte, error) {
	s := [][]byte{nil}
	for c, esc := e.next(); c != 0; c, esc = e.next() {
		l := len(s) - 1
		switch {
		case esc:
			if bcount > 0 {
				s[l] = append(s[l], c)
			} else {
				for i := range s {
					s[i] = append(s[i], c)
				}
			}
		case c == ',':
			if bcount > 0 {
				s = append(s, nil)
			} else {
				for i := range s {
					s[i] = append(s[i], c)
				}
			}
		case c == '{':
			ss, err := e.parse(bcount + 1)
			if err != nil {
				return nil, err
			}
			x := combine(s, ss)
			log.Printf("combine(%s,%s) = %s", s, ss, x)
			s = x
		case c == '}':
			if bcount <= 0 {
				return nil, fmt.Errorf("invalid expansion: %s: unbalanced bracets", e.str)
			}
			return s, nil
		default:
			if bcount > 0 {
				s[l] = append(s[l], c)
			} else {
				for i := range s {
					s[i] = append(s[i], c)
				}
			}
		}
	}
	if bcount > 0 {
		return nil, fmt.Errorf("invalid expansion: %s: unbalanced bracets", e.str)
	}
	return s, nil
}

func combine(a, b [][]byte) [][]byte {
	var res [][]byte
	for _, astr := range a {
		for _, bstr := range b {
			log.Printf("res = %s", res)
			comb := append(astr, bstr...)
			log.Printf("comb = %s", comb)
			res = append(res, comb...)
			log.Printf("%s (%s %s)", res, astr, bstr)
		}
	}
	return res
}

// func combine(a, b []string) []string {
// 	if len(a[len(a)-1]) == 0 {
// 		a = a[:len(a)-1]
// 		return append(a, b...)
// 	}
// 	var res []string
// 	for _, astr := range a {
// 		for _, bstr := range b {
// 			res = append(res, astr+bstr)
// 		}
// 	}
// 	return res
// }
