package schema

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

type modifierTestCase struct {
	input      []string
	args       []string
	shouldFail bool
	output     []string
}

func (tc *modifierTestCase) test(f VariableModifierFunction, t *testing.T) {
	t.Run(fmt.Sprintf("%s(%s)", strings.Join(tc.input, "/"), strings.Join(tc.args, "/")), func(t *testing.T) {
		res, err := f(tc.input, tc.args)
		if !tc.shouldFail {
			if err != nil {
				t.Fatalf("Modifier failed: %v", err)
			}
			if !slices.Equal(res, tc.output) {
				t.Fatalf("Incorrect modifier application result: expected \"%s\", got \"%s\"",
					strings.Join(tc.output, "/"), strings.Join(res, "/"))
			}
		} else { // should fail
			if err == nil {
				t.Fatalf("Modifier did not fail when it was expected to")
			}
		}
	})
}

func TestStripLastPrefix(t *testing.T) {
	cases := []modifierTestCase{
		{
			input:      []string{"prefix-a"},
			args:       []string{"prefix-"},
			shouldFail: false,
			output:     []string{"a"},
		},
		{
			input:      []string{"other", "prefix-a"},
			args:       []string{"prefix-"},
			shouldFail: false,
			output:     []string{"other", "a"},
		},
		{
			input:      []string{"other", "prefix-a"},
			args:       []string{"prefix-"},
			shouldFail: false,
			output:     []string{"other", "a"},
		},
		{
			input:      []string{"other", "prefix-a"},
			args:       []string{"prefix-"},
			shouldFail: false,
			output:     []string{"other", "a"},
		},
		{
			input:      []string{"a"},
			args:       []string{"prefix-"},
			shouldFail: false,
			output:     []string{"a"},
		},
		{
			input:      []string{"other", "a"},
			args:       []string{"prefix-"},
			shouldFail: false,
			output:     []string{"other", "a"},
		},
		{
			input:      []string{},
			args:       []string{"prefix-"},
			shouldFail: false,
			output:     []string{},
		},

		{
			input:      []string{"a"},
			args:       []string{},
			shouldFail: true,
		},
	}
	for _, testCase := range cases {
		testCase.test(modifierStripLastPrefix, t)
	}
}
