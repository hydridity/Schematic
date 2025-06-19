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
	String() string
	ExtractContext(input string) (ExtractedContext, error)
}

type Impl struct {
	Constraints []Constraint
	ast         *parser.SchemaAST
}

type ExtractedContext map[string]string

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
	mergedModifiers := getPredefinedModifiers()
	for k, v := range context.VariableModifiers {
		mergedModifiers[k] = v
	}

	mergedContext := ValidationContext{
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

// ExtractContext walks the schema from right to left and extracts variable value from input.
// WARNING: VERY EXPERIMENTAL PROTOTYPE CODE!!!!!!!!!!!
func (s *Impl) ExtractContext(input string) (ExtractedContext, error) {
	inputSegments := strings.Split(strings.Trim(input, "/"), "/")
	constraints := s.Constraints

	context := make(ExtractedContext)
	i := len(inputSegments) - 1
	j := len(constraints) - 1

	for j >= 0 && i >= 0 {
		switch c := constraints[j].(type) {
		case *WildcardSingleConstraint:
			// Wildcard: skip one segment
			i--
			j--
		case *VariableSetConstraint:
			// VariableSet: ignore for now
			// TODO: Should it validate against the variable store and extract only if variable store contains correct value ? same as LiteralConstraint
			i--
			j--
		case *WildcardMultiConstraint:
			// Imposible at current implementation, unable to determine the greed
			return nil, fmt.Errorf("wildcard constraint at %d cannot be used in ExtractContext", j)
		case *LiteralConstraint:
			// Literal: must match input, and skip
			if inputSegments[i] != c.Literal {
				return nil, fmt.Errorf("literal mismatch at segment %d: expected %s, got %s", i, c.Literal, inputSegments[i])
			}
			i--
			j--
		case *VariableConstraint:
			// Variable: greedily consume all remaining segments
			// TODO, perhaps handle previous literal segment as terminator for greediness as version 1 ? this needs to be indicated through modifier so we can determine the amount of wildards vs variable beginning
			// Schema: $variable_name/+/+ and input: segment1/segment2/segment3/segment4/segment5 should return context
			// if variable is greedy:
			//   extracted variable_name: segment1/segment2/segment3
			// if we know variable starts with segment2:
			//   extracted variable_name: segment2/segment3 - would be context, but this schema should error because there would need to be a literal or wildcard before variable constraint
			// if Schema is segment1/$variable_name/+/+ or +/$variable_name/+/+ and we know that variable starts with segment2:
			//   extracted variable_name: segment2/segment3
			if j != 0 {
				return nil, fmt.Errorf("durin extraction of context, variable constraints must be the leftmost constraint, found at %d", j)
			}

			segments := inputSegments[0 : i+1]
			context[c.VariableName] = strings.Join(segments, "/")
			i = -1
			j = -1
		default:
			return nil, fmt.Errorf("unsupported constraint type at %d", j)
		}
	}

	if i >= 0 || j >= 0 { // TODO: Because there is no indication on how to terminate variable greediness, we only know if we have enough segments
		return nil, fmt.Errorf("input and schema did not align: remaining input=%v, remaining constraints=%d", inputSegments[:i+1], j+1)
	}
	return context, nil
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
