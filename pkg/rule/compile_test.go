package rule

import "testing"

func TestCompile(t *testing.T) {
	tests := []struct {
		test  string
		want  float64
		iserr bool
	}{
		{"2+3", 5, false},
		{"2+true", 0, true},
		{"2+3+1", 6, false},
		{"2*3+1", 7, false},
		{"2+3*3", 11, false},
		{"2-3*3", -7, false},
		{"2-3/3+1", 2, false},
		{"2/3", 2.0 / 3.0, false},
		{"2/3>1/2", 1, false},
		{"2/3<1/2", 0, false},
		{"2/3=1/2", 0, false},
		{"2/4=0.5", 1, false},
		{"2/4>0.5", 0, false},
		{"2/4<0.5", 0, false},
	}

	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			rule, err := Compile(tc.test)
			if tc.iserr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			// t.Logf("rule = %s", rule)
			if got := rule.Execute(); got != tc.want {
				t.Fatalf("expected %f; got %f", tc.want, got)
			}
		})
	}
}
