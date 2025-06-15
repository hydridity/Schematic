package schema

import (
	"fmt"
	"strings"

	"github.com/hydridity/Schematic/pkg/parser"
	"github.com/hydridity/Schematic/pkg/schema/constraints"
	ctx "github.com/hydridity/Schematic/pkg/schema/context"
)

type Schema interface {
	Validate(input string, context *ctx.ValidationContext) error
	String() string
}

type Impl struct {
	Constraints []constraints.Constraint
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

func (s *Impl) Validate(input string, context *ctx.ValidationContext) error {
	mergedModifiers := getPredefinedModifiers()
	for k, v := range context.VariableModifiers {
		mergedModifiers[k] = v
	}

	mergedContext := ctx.ValidationContext{
		VariableStore:     context.VariableStore,
		VariableModifiers: mergedModifiers,
	}

	inputSegments := strings.Split(strings.Trim(input, "/"), "/")
	for _, constraint := range s.Constraints {
		var err error
		inputSegments, err = constraint.Consume(inputSegments, &mergedContext)
		if err != nil {
			return fmt.Errorf("failed to consume input '%s' with constraint '%s': %w", input, constraint.String(), err)
		}
	}

	if len(inputSegments) > 0 {
		return fmt.Errorf("input '%s' did not fully consume all segments, remaining: %v", input, inputSegments)
	}
	return nil
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

	compiledConstraints, err := constraints.CompileConstraints(schemaAst)
	if err != nil {
		return nil, err
	}
	return &Impl{Constraints: compiledConstraints, ast: schemaAst}, nil
}
