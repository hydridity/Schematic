package schema

import (
	"errors"
	"fmt"

	"github.com/hydridity/Schematic/pkg/parser"
)

// The basic interface of our constraints.
type Constraint interface {
	Consume([]string) ([]string, error)
	Debug() string
}

// Concrete realization of our constraints.
type LiteralConstraint struct {
	Literal string
}
type WildcardSingleConstraint struct{}
type WildcardMultiConstraint struct{}
type VariableConstraint struct {
	VariableName string
}

type VariableSetConstraint struct {
	VariableName string
}

func (c *LiteralConstraint) Consume(path []string) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}
	if path[0] != c.Literal {
		return nil, fmt.Errorf("expected '%s', got '%s'", c.Literal, path[0])
	}
	return path[1:], nil
}

func (c *LiteralConstraint) Debug() string {
	return fmt.Sprintf("LiteralConstraint(%s)", c.Literal)
}

func (c *WildcardSingleConstraint) Consume(path []string) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}
	return path[1:], nil
}
func (c *WildcardSingleConstraint) Debug() string {
	return "WildcardSingleConstraint"
}

func (c *WildcardMultiConstraint) Consume(path []string) ([]string, error) {
	return []string{}, nil
}

func (c *WildcardMultiConstraint) Debug() string {
	return "WildcardMultiConstraint"
}

func (c *VariableConstraint) Consume(path []string) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}
	if path[0] != GetVariableValue(c.VariableName) {
		// Perhaps pointer ?
		// Perhaps "variable store" in the VariableConstraint object itself ?
		return nil, errors.New("invalid variable constraint value")
	}
	return path[1:], nil
}

func (c *VariableConstraint) Debug() string {
	return fmt.Sprintf("VariableConstraint(%s)", c.VariableName)
}

func (c *VariableSetConstraint) Consume(path []string) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}
	if path[0] != GetVariableValue(c.VariableName) {
		return nil, errors.New("invalid variable set constraint value")
	}
	return path[1:], nil
}

func (c *VariableSetConstraint) Debug() string {
	return fmt.Sprintf("VariableSetConstraint(%s)", c.VariableName)
}

func GetVariableValue(name string) string {
	fmt.Println("test")
	return "test"
}

func CompileSchema(schemaAst *parser.Schema, context map[string]string) []Constraint {

	constraints := make([]Constraint, 0, len(schemaAst.Parts))

	for _, part := range schemaAst.Parts {
		switch {

		case part.Var != nil:
			fmt.Println("Var:", part.Var.Name)
			constraints = append(constraints, &VariableConstraint{VariableName: part.Var.Name})
			if part.Var.Modifier != nil {
				// TODO: handle modifiers
				fmt.Printf("Modifier: %s(%s), Argument: %s\n", part.Var.Modifier.Func, part.Var.Modifier.Arg, part.Var.Modifier.Arg)
			}

		case part.Wildcard != nil:
			fmt.Println("Wildcard:", *part.Wildcard)
			if *part.Wildcard == "+" {
				constraints = append(constraints, &WildcardSingleConstraint{})
			} else if *part.Wildcard == "*" {
				constraints = append(constraints, &WildcardMultiConstraint{})
			}

		case part.VarSet != nil:
			fmt.Println("VarSet:", part.VarSet.Name)
			constraints = append(constraints, &VariableSetConstraint{VariableName: part.VarSet.Name})

		case part.Literal != nil:
			fmt.Println("Literal:", *part.Literal)
			constraints = append(constraints, &LiteralConstraint{
				Literal: *part.Literal,
			})
		}
	}

	return constraints
}
