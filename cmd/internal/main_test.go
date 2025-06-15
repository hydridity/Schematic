// go
package main

import (
	"testing"

	"github.com/hydridity/Schematic/pkg/schema"
	"github.com/hydridity/Schematic/pkg/schema/context"
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
	schemaCompiled, err := schema.CreateSchema(schemaStr)
	if err != nil {
		t.Fatalf("Error creating schema: %v", err)
	}

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
		{
			name:           "Valid input with ansible- prefix and invalid technology",
			gitlabPath:     "deployment/group1/project1/ansible-project1-backend",
			input:          "deployment/group1/project1/project1-backend/not_allowed/admin",
			expectValidate: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &testVariableStore{
				StringVariables: map[string]string{
					"gitlab_path": tc.gitlabPath,
				},
				SetVariables: map[string][]string{
					"technologies": {"postgres", "kafka"},
				},
			}
			err := schemaCompiled.Validate(tc.input, &context.ValidationContext{VariableStore: store})
			if tc.expectValidate && err != nil {
				t.Errorf("expected validation to succeed, got error: %v", err)
			}
			if !tc.expectValidate && err == nil {
				t.Errorf("expected validation to fail, but it succeeded")
			}
		})
	}
}

func TestSchemaFormat(t *testing.T) {
	tests := []struct {
		name            string
		schemaStr       string
		gitlabPath      string
		input           string
		expectValidate  bool
		expectSchemaErr bool
	}{
		{
			name:            "Valid input with valid schema and invalid context",
			schemaStr:       `$gitlab_path.strip_last_prefix("helm-", "ansible-")/$[technologies]/+`,
			gitlabPath:      "deployment/group1/project1/helm-project1-backend",
			input:           "deployment/invalid-group1/invalid-project1/project1-backend/postgres/admin",
			expectValidate:  false,
			expectSchemaErr: false,
		},
		{
			name:            "Valid input with invalid schema and valid context",
			schemaStr:       `$gitlab_path.strip_last_prefix("helm-", "ansible-")///`,
			gitlabPath:      "deployment/group1/project1/helm-project1-backend",
			input:           "deployment/group1/project1/something-project1-backend/postgres/admin",
			expectValidate:  false,
			expectSchemaErr: true,
		},
		{
			name:            "Valid input with ansible- prefix",
			schemaStr:       `$gitlab_path.strip_last_prefix("helm-", "ansible-")/$[technologies]/+`,
			gitlabPath:      "deployment/group1/project1/ansible-project1-backend",
			input:           "deployment/group1/project1/project1-backend/postgres/admin",
			expectValidate:  true,
			expectSchemaErr: false,
		},
		{
			name:            "Completely invalid schema",
			schemaStr:       `///`,
			gitlabPath:      "deployment/group1/project1/helm-project1-backend",
			input:           "deployment/group1/project1/project1-backend/postgres/admin",
			expectValidate:  false,
			expectSchemaErr: true,
		},
		{
			name:            "Empty schema",
			schemaStr:       ``,
			gitlabPath:      "deployment/group1/project1/helm-project1-backend",
			input:           "deployment/group1/project1/project1-backend/postgres/admin",
			expectValidate:  false,
			expectSchemaErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &testVariableStore{
				StringVariables: map[string]string{
					"gitlab_path": tc.gitlabPath,
				},
				SetVariables: map[string][]string{
					"technologies": {"postgres", "kafka"},
				},
			}
			schemaCompiled, err := schema.CreateSchema(tc.schemaStr)
			if tc.expectSchemaErr {
				if err == nil {
					t.Errorf("expected schema compilation error, but got none")
				}
				// If schema compilation fails, skip validation
				return
			} else {
				if err != nil {
					t.Fatalf("Error creating schema: %v", err)
				}
			}

			err = schemaCompiled.Validate(tc.input, &context.ValidationContext{VariableStore: store})
			if tc.expectValidate && err != nil {
				t.Errorf("expected validation to succeed, got error: %v", err)
			}
			if !tc.expectValidate && err == nil {
				t.Errorf("expected validation to fail, but it succeeded")
			}
		})
	}
}
