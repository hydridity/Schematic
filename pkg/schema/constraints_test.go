package schema

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

type constraintTestCase struct {
	constraint   Constraint
	path         []string
	shouldFail   bool
	expectedRest []string

	variablesMap         map[string]string
	allowedVariableNames map[string]struct{}

	setsMap         map[string][]string
	allowedSetNames map[string]struct{}
}

type testVariableStore struct {
	t        *testing.T
	testCase *constraintTestCase
}

func (vs *testVariableStore) GetVariable(name string) (string, bool) {
	_, isAllowed := vs.testCase.allowedVariableNames[name]
	if !isAllowed {
		vs.t.Fatalf("Requested variable %s, which isn't allowed.", name)
	}
	value, found := vs.testCase.variablesMap[name]
	return value, found
}

func (vs *testVariableStore) GetVariableSet(name string) ([]string, bool) {
	_, isAllowed := vs.testCase.allowedSetNames[name]
	if !isAllowed {
		vs.t.Fatalf("Requested variable set %s, which isn't allowed.", name)
	}
	value, found := vs.testCase.setsMap[name]
	return value, found
}

func (tc *constraintTestCase) test(t *testing.T) {
	t.Run(fmt.Sprintf("%s", strings.Join(tc.path, ":")), func(t *testing.T) {
		ctx := ValidationContext{
			VariableStore:     &testVariableStore{t: t, testCase: tc},
			VariableModifiers: getPredefinedModifiers(),
		}
		rest, err := tc.constraint.Consume(tc.path, &ctx)
		if !tc.shouldFail {
			if err != nil {
				t.Fatalf("Constraint failed when it was expected to succeed: %v", err)
			}
			if !slices.Equal(rest, tc.expectedRest) {
				t.Fatalf("Constraint must return \"%s\", instead it was \"%s\".",
					strings.Join(tc.expectedRest, "/"), strings.Join(rest, "/"))
			}
		} else { // should fail
			if err == nil {
				t.Fatalf("Constraint succeeded when it was expected to fail")
			}
		}
	})
}

func TestLiteralConstraint(t *testing.T) {
	c := func(value string) Constraint {
		return &LiteralConstraint{value}
	}

	cases := []constraintTestCase{
		{
			constraint:   c("hello"),
			path:         []string{"hello"},
			shouldFail:   false,
			expectedRest: []string{},
		},
		{
			constraint:   c("hello"),
			path:         []string{"hello", "world"},
			shouldFail:   false,
			expectedRest: []string{"world"},
		},
		{
			constraint:   c("hello"),
			path:         []string{"hello", "hello"},
			shouldFail:   false,
			expectedRest: []string{"hello"},
		},
		{
			constraint:   c("hello"),
			path:         []string{"hello", "a", "b", "c"},
			shouldFail:   false,
			expectedRest: []string{"a", "b", "c"},
		},

		{
			constraint:   c("hello"),
			path:         []string{},
			shouldFail:   true,
			expectedRest: nil,
		},
		{
			constraint:   c("hello"),
			path:         []string{"world"},
			shouldFail:   true,
			expectedRest: nil,
		},
	}
	for _, testCase := range cases {
		testCase.test(t)
	}
}

func TestWildcardSingleConstraint(t *testing.T) {
	c := func(minSegments int, maxSegments int) Constraint {
		return &WildcardSingleConstraint{minSegments, maxSegments}
	}
	// TODO: add different min/max test cases when the behaviour is finalized
	cases := []constraintTestCase{
		{
			constraint:   c(1, 1),
			path:         []string{"a"},
			shouldFail:   false,
			expectedRest: []string{},
		},
		{
			constraint:   c(1, 1),
			path:         []string{"a", "b"},
			shouldFail:   false,
			expectedRest: []string{"b"},
		},
		{
			constraint:   c(1, 1),
			path:         []string{"hello"},
			shouldFail:   false,
			expectedRest: []string{},
		},
		{
			constraint:   c(1, 1),
			path:         []string{"hello", "world"},
			shouldFail:   false,
			expectedRest: []string{"world"},
		},
		{
			constraint:   c(1, 1),
			path:         []string{"a", "b", "c", "d"},
			shouldFail:   false,
			expectedRest: []string{"b", "c", "d"},
		},

		{
			constraint:   c(1, 1),
			path:         []string{},
			shouldFail:   true,
			expectedRest: nil,
		},
	}
	for _, testCase := range cases {
		testCase.test(t)
	}
}

func TestWildcardMultiConstraint(t *testing.T) {
	c := &WildcardMultiConstraint{}
	cases := []constraintTestCase{
		{
			constraint:   c,
			path:         []string{},
			shouldFail:   false,
			expectedRest: []string{},
		},
		{
			constraint:   c,
			path:         []string{"a"},
			shouldFail:   false,
			expectedRest: []string{},
		},
		{
			constraint:   c,
			path:         []string{"a", "b", "c", "d"},
			shouldFail:   false,
			expectedRest: []string{},
		},
		{
			constraint:   c,
			path:         []string{"hello"},
			shouldFail:   false,
			expectedRest: []string{},
		},
		{
			constraint:   c,
			path:         []string{"hello", "world"},
			shouldFail:   false,
			expectedRest: []string{},
		},
	}
	for _, testCase := range cases {
		testCase.test(t)
	}
}

func TestVariableConstraint(t *testing.T) {
	t.Run("EnsureCorrectVariableName", func(t *testing.T) {
		expectedVarName := "varName"
		constraint := &VariableConstraint{VariableName: expectedVarName}
		resVarName := constraint.GetVariableName()
		if resVarName != expectedVarName {
			t.Fatalf("Expected the variable name to be \"%s\", got \"%s\"", expectedVarName, resVarName)
		}
	})

	t.Run("NoModifiers", func(t *testing.T) {
		c := func(varName string) Constraint {
			return &VariableConstraint{VariableName: varName}
		}
		cases := []constraintTestCase{
			{
				constraint:   c("varName"),
				path:         []string{"value"},
				shouldFail:   false,
				expectedRest: []string{},
				variablesMap: map[string]string{
					"varName": "value",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
			{
				constraint:   c("varName"),
				path:         []string{"value", "rest"},
				shouldFail:   false,
				expectedRest: []string{"rest"},
				variablesMap: map[string]string{
					"varName": "value",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
			{
				constraint:   c("varName"),
				path:         []string{"a", "b", "c"},
				shouldFail:   false,
				expectedRest: []string{"c"},
				variablesMap: map[string]string{
					"varName": "a/b",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},

			{
				constraint:   c("varName"),
				path:         []string{"other"},
				shouldFail:   true,
				expectedRest: nil,
				variablesMap: map[string]string{
					"varName": "value",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
			{
				constraint:   c("varName"),
				path:         []string{},
				shouldFail:   true,
				expectedRest: nil,
				variablesMap: map[string]string{
					"varName": "value",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
			{
				constraint:   c("varName"),
				path:         []string{"something"},
				shouldFail:   true,
				expectedRest: nil,
				variablesMap: map[string]string{},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
		}
		for _, testCase := range cases {
			testCase.test(t)
		}
	})

	t.Run("WithSingleModifier", func(t *testing.T) {
		c := &VariableConstraint{VariableName: "varName", Modifiers: []VariableModifier{{
			FuncName: "strip_last_prefix",
			Args:     []string{"prefix-"},
		}}}
		cases := []constraintTestCase{
			{
				constraint:   c,
				path:         []string{"hello"},
				shouldFail:   false,
				expectedRest: []string{},
				variablesMap: map[string]string{
					"varName": "prefix-hello",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
			{
				constraint:   c,
				path:         []string{"a", "b", "c", "d"},
				shouldFail:   false,
				expectedRest: []string{"d"},
				variablesMap: map[string]string{
					"varName": "a/b/prefix-c",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},

			{
				constraint:   c,
				path:         []string{"x"},
				shouldFail:   true,
				expectedRest: nil,
				variablesMap: map[string]string{
					"varName": "prefix-hello",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
			{
				constraint:   c,
				path:         []string{"prefix-hello"},
				shouldFail:   true,
				expectedRest: nil,
				variablesMap: map[string]string{
					"varName": "prefix-hello",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
			{
				constraint:   c,
				path:         []string{"other", "hello"},
				shouldFail:   true,
				expectedRest: []string{},
				variablesMap: map[string]string{
					"varName": "prefix-hello",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
			{
				constraint:   c,
				path:         []string{"other", "hello"},
				shouldFail:   true,
				expectedRest: []string{},
				variablesMap: map[string]string{
					"varName": "a/prefix-hello",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
			{
				constraint:   c,
				path:         []string{"a", "other", "hello"},
				shouldFail:   true,
				expectedRest: []string{},
				variablesMap: map[string]string{
					"varName": "prefix-hello",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
		}
		for _, testCase := range cases {
			testCase.test(t)
		}
	})

	t.Run("WithTwoModifiers", func(t *testing.T) {
		c := &VariableConstraint{VariableName: "varName", Modifiers: []VariableModifier{
			{
				FuncName: "strip_last_prefix",
				Args:     []string{"first-"},
			},
			{
				FuncName: "strip_last_prefix",
				Args:     []string{"second-"},
			},
		}}
		cases := []constraintTestCase{
			{
				constraint:   c,
				path:         []string{"hello"},
				shouldFail:   false,
				expectedRest: []string{},
				variablesMap: map[string]string{
					"varName": "first-second-hello",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
			{
				constraint:   c,
				path:         []string{"x", "hello"},
				shouldFail:   false,
				expectedRest: []string{},
				variablesMap: map[string]string{
					"varName": "x/first-second-hello",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},

			{
				constraint:   c,
				path:         []string{"hello"},
				shouldFail:   true,
				expectedRest: nil,
				variablesMap: map[string]string{
					"varName": "second-first-hello",
				},
				allowedVariableNames: map[string]struct{}{
					"varName": {},
				},
			},
		}
		for _, testCase := range cases {
			testCase.test(t)
		}
	})
}
