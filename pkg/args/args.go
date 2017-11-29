package args

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// IntList is a slice of integers.
type IntList []int

// String satisfies the flag.Value interface for IntList.
func (is IntList) String() string {
	var strs []string
	for _, i := range is {
		strs = append(strs, fmt.Sprintf("%d", i))
	}
	return strings.Join(strs, ",")
}

// Set satisfies the flag.Value interface for IntList.
func (is *IntList) Set(val string) error {
	l, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	*is = append(*is, l)
	return nil
}

// StringList is a slice of strings.
type StringList []string

// String satisfies the flag.Value interface for StringList.
func (ss StringList) String() string {
	return strings.Join(ss, ",")
}

// Set satisfies the flag.Value interface for StringList.
func (ss *StringList) Set(val string) error {
	*ss = append(*ss, val)
	return nil
}

// RegexList is a slice of regular expressions.
type RegexList []*regexp.Regexp

// String satisfies the flag.Value interface for RegexList.
func (rs RegexList) String() string {
	var strs []string
	for _, r := range rs {
		strs = append(strs, r.String())
	}
	return strings.Join(strs, ",")
}

// Set satisfies the flag.Value interface for RegexList.
func (rs *RegexList) Set(val string) error {
	r, err := regexp.Compile(val)
	if err != nil {
		return err
	}
	*rs = append(*rs, r)
	return nil
}
