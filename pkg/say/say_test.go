package say_test

import (
	"bytes"
	"regexp"
	"testing"

	"bitbucket.org/fflo/semix/pkg/say"
)

func Test(t *testing.T) {
	tests := []struct {
		test, re string
	}{
		{"hello", ".* hello"},
		{"this is my message", ".* this is my message"},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			say.SetColor(false)
			say.SetDebug(false)
			output := new(bytes.Buffer)
			say.SetOutput(output)
			say.Info(tc.test)
			if got := output.String(); !regexp.MustCompile(tc.re).MatchString(got) {
				t.Fatalf("expected %s; got %s", tc.re, got)
			}
			output = new(bytes.Buffer)
			say.SetOutput(output)
			say.Debug(tc.test)
			if got := output.String(); got != "" {
				t.Fatalf("got %s", got)
			}
			say.SetDebug(true)
			output = new(bytes.Buffer)
			say.SetOutput(output)
			say.Debug(tc.test)
			if got := output.String(); !regexp.MustCompile(tc.re).MatchString(got) {
				t.Fatalf("expected %s; got %s", tc.re, got)
			}
			say.SetColor(true)
			output = new(bytes.Buffer)
			say.SetOutput(output)
			say.Debug(tc.test)
			if got := output.String(); !regexp.MustCompile(tc.re).MatchString(got) {
				t.Fatalf("expected %s; got %s", tc.re, got)
			}
		})
	}
}
