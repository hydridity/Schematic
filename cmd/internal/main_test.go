// go
package main

import (
	"testing"
	"github.com/hydridity/Schematic/pkg/parser"
	"github.com/hydridity/Schematic/pkg/schema"
)

type testVariableStore struct {
	StringVariables map[string]string
	SetVariables    map[string][]string
}

func (vs *testVariableStore) GetVariable(name string) (string, bool) {
	v, ok := vs.StringVariables[name]
	return v, ok
}

func (vs *testVariableStore) GetVariableSet(name string) ([]string, bool) {
	v, ok := vs.SetVariables[name]
	return v, ok
}

func TestVariableModifiers(t *testing.T) {
	schemaStr := `$gitlab_path.strip_last_prefix("helm-", "ansible-")/$[technologies]/+`
	parser, err := parser.NewParser()
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}
	schemaAst, err := parser.ParseString("", schemaStr)
	if err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}
	schemaCompiled := schema.CompileSchema(schemaAst, nil)

	tests := []struct {
		name           string
		gitlabPath     string
		input          string
		expectValidate bool
	}{
		{
			name:           "Valid input with helm- prefix",
			gitlabPath:     "deployment/group1/project1/helm-project1-backend",
			input:          "deployment/group1/project1/project1-backend/postgres/admin",
			expectValidate: true,
		},
		{
			name:           "Invalid input with something- prefix",
			gitlabPath:     "deployment/group1/project1/helm-project1-backend",
			input:          "deployment/group1/project1/something-project1-backend/postgres/admin",
			expectValidate: false,
		},
		{
			name:           "Valid input with ansible- prefix",
			gitlabPath:     "deployment/group1/project1/ansible-project1-backend",
			input:          "deployment/group1/project1/project1-backend/postgres/admin",
			expectValidate: true,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			store := &testVariableStore{
				StringVariables: map[string]string{
					"gitlab_path": tc.gitlabPath,
				},
				SetVariables: map[string][]string{
					"technologies": {"postgres", "kafka"},
				},
			}
			err := Validate(tc.input, schemaCompiled, store)
			if tc.expectValidate && err != nil {
				t.Errorf("expected validation to succeed, got error: %v", err)
			}
			if !tc.expectValidate && err == nil {
				t.Errorf("expected validation to fail, but it succeeded")
			}
		})
	}
}