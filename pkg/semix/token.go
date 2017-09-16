package semix

import (
	"fmt"
)

// Token denotes a  token in an input document. It holds the according Concept
// or nil and its position in the input document.
type Token struct {
	Token      string
	Concept    *Concept
	Begin, End int
}

// String returns the string representation of a token.
func (t Token) String() string {
	return fmt.Sprintf("%q %q %d %d", t.Token, t.Concept, t.Begin, t.End)
}
