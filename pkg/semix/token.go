package semix

import (
	"fmt"
)

// Token denotes a  token in an input document. It holds the according Concept
// or nil and its position in the input document.
type Token struct {
	Token, Path string
	Concept     *Concept
	Begin, End  int
}

// String returns the string representation of a token.
func (t Token) String() string {
	return fmt.Sprintf("%q %q %d %d %q", t.Token, t.Concept, t.Begin, t.End, t.Path)
}
