package schema

import (
	"fmt"
	"strings"

	"github.com/hydridity/Schematic/pkg/parser"
)

type VariableStore interface {
	GetVariable(name string) (string, bool)
	GetVariableSet(name string) ([]string, bool)
}

// VariableModifierFunction represents a Modifier. It accepts a context variable value, split by '/', along with
// a set of schema-provided arguments, and should modify the variable however it wants.
// It should return the modified variable slice.
type VariableModifierFunction func(variable []string, args []string) ([]string, error)

type ValidationContext struct {
	VariableStore     VariableStore
	VariableModifiers map[string]VariableModifierFunction
}

type Schema interface {
	Validate(input string, context *ValidationContext) error
	consume(inputSegments []string, context *ValidationContext) ([]string, error)
	String() string
}

type Impl struct {
	Constraints []Constraint
	ast         *parser.SchemaAST
}

func (s *Impl) String() string {
	builder := strings.Builder{}
	builder.WriteString("AST:\n")
	builder.WriteString(s.ast.String())
	builder.WriteString("\n\nConstraints:\n")
	for _, constraint := range s.Constraints {
		builder.WriteString(constraint.String())
		builder.WriteString("\n")
	}
	return builder.String()
}

func (s *Impl) Validate(input string, context *ValidationContext) error {
	inputSegments := strings.Split(strings.Trim(input, "/"), "/")

	remainingSegments, err := s.consume(inputSegments, context)
	if err != nil {
		return err
	}

	if len(remainingSegments) > 0 {
		return fmt.Errorf("input '%s' did not fully consume all segments, remaining: %v", input, inputSegments)
	}
	return nil
}

func (s *Impl) consume(inputSegments []string, context *ValidationContext) ([]string, error) {
	mergedModifiers := getPredefinedModifiers()
	for k, v := range context.VariableModifiers {
		mergedModifiers[k] = v
	}

	mergedContext := ValidationContext{
		VariableStore:     context.VariableStore,
		VariableModifiers: mergedModifiers,
	}

	for _, constraint := range s.Constraints {
		var err error
		inputSegments, err = constraint.Consume(inputSegments, &mergedContext)
		if err != nil {
			return nil, fmt.Errorf("failed to consume input '%s' with constraint %w", constraint.String(), err)
		}
	}
	return inputSegments, nil
}

func CreateSchema(schemaStr string) (Schema, error) {
	parserObj, err := parser.NewParser()
	if err != nil {
		return nil, err
	}

	schemaAst, err := parserObj.ParseString("", schemaStr)
	if err != nil {
		return nil, err
	}

	constraints := CompileConstraints(schemaAst)
	return &Impl{Constraints: constraints, ast: schemaAst}, nil
}
