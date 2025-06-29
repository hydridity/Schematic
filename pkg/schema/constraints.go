package schema

import (
	"errors"
	"fmt"
	"regexp"
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

type RegexConstraint struct {
	Pattern string
}

type WildcardSingleConstraint struct {
	Min int
	Max int
}
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

func (c *RegexConstraint) Consume(path []string, context *ValidationContext) ([]string, error) {
	if len(path) <= 0 {
		return nil, errors.New("empty path")
	}

	pattern, err := regexp.Compile(c.Pattern)
	if err != nil {
		return nil, err
	}

	if !pattern.MatchString(path[0]) {
		return nil, fmt.Errorf("expected '%s', does not match regex '%s'", path[0], c.Pattern)
	}
	return path[1:], nil
}

func (c *RegexConstraint) String() string {
	return fmt.Sprintf("RegexConstraint(%s)", c.Pattern)
}

func (c *RegexConstraint) GetVariableName() string {
	return "" // Regex constraints do not have a variable name
}

func (c *WildcardSingleConstraint) Consume(path []string, context *ValidationContext) ([]string, error) {
	// if len(path) <= 0 {
	// 	return nil, errors.New("empty path")
	// }
	//schema: literal1/+{0,2}
	//path [literal1/test]
	if len(path) < c.Min {
		return nil, errors.New(fmt.Sprintf("not enough segments for quantified wildcard: need at least %d", c.Min))
	}

	if len(path) >= c.Max {
		return path[c.Max:], nil
	}

	if len(path) >= c.Min && len(path) <= c.Max {
		return []string{}, nil
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
		subSchemaInput := make([]string, len(path))
		copy(subSchemaInput, path)
		subSchemaCompiled, err := CreateSchema(v)
		if err != nil {
			return nil, err
		}

		subSchemaInput, err = subSchemaCompiled.consume(subSchemaInput, context)

		if err == nil {
			foundInSet = true
			path = subSchemaInput
			break
		}
	}

	if !foundInSet {
		return nil, errors.New("invalid variable set constraint value")
	}

	return path, nil
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
				if part.Wildcard.Quantifier != nil {
					//TODO: Do we need checks for individual fields ?
					min := part.Wildcard.Quantifier.Min
					max := part.Wildcard.Quantifier.Max
					constraints = append(constraints, &WildcardSingleConstraint{Min: min, Max: max})
				} else {
					//We handle behaviour as before
					constraints = append(constraints, &WildcardSingleConstraint{Min: 1, Max: 1})
				}
			} else if part.Wildcard.Symbol == "*" {
				constraints = append(constraints, &WildcardMultiConstraint{})
			}

		case part.VarSet != nil:
			constraints = append(constraints, &VariableSetConstraint{VariableName: part.VarSet.Name})

		case part.Literal != nil:
			constraints = append(constraints, &LiteralConstraint{
				Literal: *part.Literal,
			})
		case part.Regex != nil:
			constraints = append(constraints, &RegexConstraint{Pattern: *part.Regex})
		}
	}

	return constraints
}
