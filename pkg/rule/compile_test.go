package rule

import "testing"

func TestCompile(t *testing.T) {
	tests := []struct {
		test string
		want float64
	}{
		{"2+3", 5},
		{"2+3+1", 6},
		{"2*3+1", 7},
		{"2+3*3", 11},
		{"2-3*3", -7},
		{"2-3/3+1", 2},
		{"2/3", 2.0 / 3.0},
		{"2/3>1/2", 1},
		{"2/3<1/2", 0},
		{"2/3=1/2", 0},
		{"2/4=0.5", 1},
		{"2/4>0.5", 0},
		{"2/4<0.5", 0},
	}

	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			rule, err := Compile(tc.test)
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
