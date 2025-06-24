package schema

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hydridity/Schematic/pkg/parser"
)

// The basic interface of our constraints.
type Constraint interface {
	Consume([]string, *ValidationContext) ([]string, error)
	String() string
	GetVariableName() string
}

// Concrete realization of our constraints.
type LiteralConstraint struct {
	Literal string
}
type WildcardSingleConstraint struct{}
type WildcardMultiConstraint struct{}

type VariableModifier struct {
	FuncName string
	Args     []string
}
type VariableModifierInstance struct {
	Modifier VariableModifier
	Function VariableModifierFunction
}
type VariableConstraint struct {
	VariableName string
	Modifiers    []VariableModifier
}

type VariableSetConstraint struct {
	VariableName string
}

func (c *LiteralConstraint) Consume(path []string, context *ValidationContext) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}
	if path[0] != c.Literal {
		return nil, fmt.Errorf("expected '%s', got '%s'", c.Literal, path[0])
	}
	return path[1:], nil
}

func (c *LiteralConstraint) String() string {
	return fmt.Sprintf("LiteralConstraint(%s)", c.Literal)
}

func (c *LiteralConstraint) GetVariableName() string {
	return "" // Literal constraints do not have a variable name
}

func (c *WildcardSingleConstraint) Consume(path []string, context *ValidationContext) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}
	return path[1:], nil
}
func (c *WildcardSingleConstraint) String() string {
	return "WildcardSingleConstraint"
}

func (c *WildcardSingleConstraint) GetVariableName() string {
	return ""
}

func (c *WildcardMultiConstraint) Consume(path []string, context *ValidationContext) ([]string, error) {
	return []string{}, nil
}

func (c *WildcardMultiConstraint) String() string {
	return "WildcardMultiConstraint"
}

func (c *WildcardMultiConstraint) GetVariableName() string {
	return ""
}

func (c *VariableConstraint) Consume(path []string, context *ValidationContext) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}
	variable, found := context.VariableStore.GetVariable(c.VariableName)
	if !found {
		return nil, fmt.Errorf("variable '%s' not found in store", c.VariableName)
	}

	// Get modifier functions referenced in the constraint
	modifierInstances := make([]VariableModifierInstance, 0, len(c.Modifiers))
	for _, modifier := range c.Modifiers {
		fun, found := context.VariableModifiers[modifier.FuncName]
		if !found {
			return nil, fmt.Errorf("modifier '%s' not found in context modifiers", modifier.FuncName)
		}
		modifierInstances = append(modifierInstances, VariableModifierInstance{
			Modifier: modifier,
			Function: fun,
		})
	}

	// Apply the modifier functions to the variable in order
	var err error
	parts := strings.Split(variable, "/")
	for _, mod := range modifierInstances {
		parts, err = mod.Function(parts, mod.Modifier.Args)
		if err != nil {
			return nil, fmt.Errorf("modifier '%s' application failed: %v", mod.Modifier.FuncName, err)
		}
	}

	// Validate the input with modified variable parts

	if len(path) < len(parts) {
		return nil, fmt.Errorf("path too short for variable '%s'", variable)
	}
	for i, part := range parts {
		if path[i] != part {
			return nil, fmt.Errorf("invalid variable constraint value at part %d, variable '%s'", i, variable)
		}
	}
	return path[len(parts):], nil
}

func (c *VariableConstraint) String() string {
	return fmt.Sprintf("VariableConstraint(%s)", c.VariableName)
}

func (c *VariableConstraint) GetVariableName() string {
	return c.VariableName
}

func (c *VariableSetConstraint) Consume(path []string, context *ValidationContext) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}
	variable, found := context.VariableStore.GetVariableSet(c.VariableName)
	if !found {
		return nil, fmt.Errorf("variable '%s' not found in store", c.VariableName)
	}
	if len(variable) == 0 {
		return nil, fmt.Errorf("variable set '%s' is empty", c.VariableName)
	}
	foundInSet := false
	for _, v := range variable {
		if path[0] == v {
			foundInSet = true
			break
		}
	}
	if !foundInSet {
		return nil, errors.New("invalid variable set constraint value")
	}
	return path[1:], nil
}

func (c *VariableSetConstraint) String() string {
	return fmt.Sprintf("VariableSetConstraint(%s)", c.VariableName)
}

func (c *VariableSetConstraint) GetVariableName() string {
	return c.VariableName
}

func CompileConstraints(schemaAst *parser.SchemaAST) []Constraint {
	constraints := make([]Constraint, 0, len(schemaAst.Parts))

	for _, part := range schemaAst.Parts {
		switch {

		case part.Var != nil:
			modifiers := make([]VariableModifier, 0)
			if part.Var.Modifier != nil {
				modifiers = append(modifiers, VariableModifier{
					FuncName: part.Var.Modifier.Func,
					Args:     part.Var.Modifier.Args,
				})
			}
			constraints = append(constraints, &VariableConstraint{VariableName: part.Var.Name, Modifiers: modifiers})

		case part.Wildcard != nil:
			if part.Wildcard.Symbol == "+" {
				constraints = append(constraints, &WildcardSingleConstraint{})
			} else if part.Wildcard.Symbol == "*" {
				constraints = append(constraints, &WildcardMultiConstraint{})
			}

		case part.VarSet != nil:
			constraints = append(constraints, &VariableSetConstraint{VariableName: part.VarSet.Name})

		case part.Literal != nil:
			constraints = append(constraints, &LiteralConstraint{
				Literal: *part.Literal,
			})
		}
	}

	return constraints
}
